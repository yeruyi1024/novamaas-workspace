package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type RequestError struct {
	StatusCode int
	Err        error
}

func (err *RequestError) Error() string {
	if err == nil || err.Err == nil {
		return "storage request failed"
	}
	return err.Err.Error()
}

func (err *RequestError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func (err *RequestError) HTTPStatusCode() int {
	if err == nil || err.StatusCode == 0 {
		return http.StatusInternalServerError
	}
	return err.StatusCode
}

type mediaCandidate struct {
	jsonPath    string
	contentType string
	payload     []byte
	extension   string
	hash        string
}

type mediaContainer struct {
	jsonPath string
	dataURI  string
}

func MaterializeVolcNativeBase64(ctx context.Context, body []byte, userID int, requestID string, taskID string, policyKey string) ([]byte, int, error) {
	if userID <= 0 || strings.TrimSpace(taskID) == "" {
		return nil, 0, &RequestError{StatusCode: http.StatusInternalServerError, Err: errors.New("base64 staging requires an authenticated user and task ID")}
	}
	if !gjson.ValidBytes(body) {
		return nil, 0, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("request body must be valid JSON")}
	}
	candidateContainers, err := findVolcNativeDataURIContainers(body)
	if err != nil {
		return nil, 0, err
	}
	if len(candidateContainers) == 0 {
		return body, 0, nil
	}
	if policyKey == "" {
		policyKey = model.StoragePolicyRelayMediaTemp
	}
	policy, profile, credential, err := ResolveEnabledPolicy(policyKey)
	if err != nil {
		return nil, 0, &RequestError{StatusCode: http.StatusServiceUnavailable, Err: fmt.Errorf("base64 staging storage is unavailable: %w", err)}
	}
	candidates, err := decodeMediaCandidates(candidateContainers, policy)
	if err != nil {
		return nil, 0, err
	}
	driver, err := driverForProfile(profile, credential)
	if err != nil {
		return nil, 0, &RequestError{StatusCode: http.StatusServiceUnavailable, Err: fmt.Errorf("initialize base64 staging storage: %w", err)}
	}

	createdObjectIDs := make([]int64, 0, len(candidates))
	materialized := append([]byte(nil), body...)
	cleanupCreatedObjects := func(message string) {
		for _, objectID := range createdObjectIDs {
			_ = model.MarkStorageObjectForDeletion(objectID, common.GetTimestamp(), message)
		}
	}
	urlsByHash := make(map[string]string, len(candidates))
	for index, candidate := range candidates {
		if signedURL, ok := urlsByHash[candidate.hash]; ok {
			materialized, err = replaceMaterializedMediaURL(materialized, candidate.jsonPath, signedURL)
			if err != nil {
				cleanupCreatedObjects("request serialization failed")
				return nil, 0, &RequestError{StatusCode: http.StatusInternalServerError, Err: fmt.Errorf("replace temporary media URL: %w", err)}
			}
			continue
		}
		objectID, objectKey, err := createRelayStorageObject(profile, policy, candidate, userID, requestID, taskID, index)
		if err != nil {
			cleanupCreatedObjects("request materialization failed")
			return nil, 0, &RequestError{StatusCode: http.StatusInternalServerError, Err: fmt.Errorf("record temporary media object: %w", err)}
		}
		createdObjectIDs = append(createdObjectIDs, objectID)
		etag, err := driver.PutObject(ctx, objectKey, candidate.contentType, int64(len(candidate.payload)), bytes.NewReader(candidate.payload))
		if err != nil {
			_ = model.MarkStorageObjectForDeletion(objectID, common.GetTimestamp(), "upload failed")
			cleanupCreatedObjects("request materialization failed")
			return nil, 0, &RequestError{StatusCode: http.StatusBadGateway, Err: fmt.Errorf("upload temporary media: %w", err)}
		}
		if err = model.MarkStorageObjectUploaded(objectID, etag); err != nil {
			_ = model.MarkStorageObjectForDeletion(objectID, common.GetTimestamp(), "persist upload state failed")
			cleanupCreatedObjects("request materialization failed")
			return nil, 0, &RequestError{StatusCode: http.StatusInternalServerError, Err: fmt.Errorf("persist temporary media upload: %w", err)}
		}
		signedURL, err := driver.PresignGet(ctx, objectKey, policySignedURLTTL(policy))
		if err != nil {
			_ = model.MarkStorageObjectForDeletion(objectID, common.GetTimestamp(), "URL signing failed")
			cleanupCreatedObjects("request materialization failed")
			return nil, 0, &RequestError{StatusCode: http.StatusBadGateway, Err: fmt.Errorf("sign temporary media URL: %w", err)}
		}
		urlsByHash[candidate.hash] = signedURL
		materialized, err = replaceMaterializedMediaURL(materialized, candidate.jsonPath, signedURL)
		if err != nil {
			cleanupCreatedObjects("request serialization failed")
			return nil, 0, &RequestError{StatusCode: http.StatusInternalServerError, Err: fmt.Errorf("replace temporary media URL: %w", err)}
		}
	}
	return materialized, len(urlsByHash), nil
}

func replaceMaterializedMediaURL(body []byte, path string, signedURL string) ([]byte, error) {
	return sjson.SetBytes(body, path, signedURL)
}

func findVolcNativeDataURIContainers(body []byte) ([]mediaContainer, error) {
	content := gjson.GetBytes(body, "content")
	if !content.IsArray() {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("content must be an array")}
	}
	containers := make([]mediaContainer, 0)
	for itemIndex, item := range content.Array() {
		if !item.IsObject() {
			continue
		}
		media := item.Get("image_url")
		if !media.Exists() {
			continue
		}
		if !media.IsObject() {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("content[%d].image_url must be an object", itemIndex)}
		}
		url := media.Get("url")
		if url.Type != gjson.String || strings.TrimSpace(url.String()) == "" {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("content[%d].image_url.url must be a non-empty string", itemIndex)}
		}
		rawURL := strings.TrimSpace(url.String())
		if strings.HasPrefix(strings.ToLower(rawURL), "data:") {
			containers = append(containers, mediaContainer{
				jsonPath: fmt.Sprintf("content.%d.image_url.url", itemIndex),
				dataURI:  rawURL,
			})
		}
	}
	return containers, nil
}

func decodeMediaCandidates(containers []mediaContainer, policy *model.StoragePolicy) ([]mediaCandidate, error) {
	if len(containers) > policy.MaxFiles {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("base64 media count exceeds the limit of %d", policy.MaxFiles)}
	}
	allowedTypes := make(map[string]struct{})
	for _, value := range strings.Split(policy.AllowedMIMETypes, ",") {
		allowedTypes[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
	}
	candidates := make([]mediaCandidate, 0, len(containers))
	var totalBytes int64
	for _, container := range containers {
		metadata, encoded, ok := strings.Cut(container.dataURI, ",")
		metadata = strings.ToLower(metadata)
		if !ok || !strings.HasSuffix(metadata, ";base64") {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("media data URI must use base64 encoding")}
		}
		contentType := strings.TrimSpace(strings.TrimSuffix(metadata[5:], ";base64"))
		if _, allowed := allowedTypes[contentType]; !allowed {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("unsupported base64 media type: %s", contentType)}
		}
		if int64(len(encoded)) > policy.MaxFileBytes*4/3+8 {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("base64 media exceeds the per-file limit of %d bytes", policy.MaxFileBytes)}
		}
		payload, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("media data URI contains invalid base64")}
		}
		if int64(len(payload)) > policy.MaxFileBytes {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("base64 media exceeds the per-file limit of %d bytes", policy.MaxFileBytes)}
		}
		actualType := http.DetectContentType(payload)
		if actualType != contentType {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("base64 media type mismatch: declared %s, detected %s", contentType, actualType)}
		}
		extension, ok := mediaExtension(contentType)
		if !ok {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("unsupported base64 media type: %s", contentType)}
		}
		totalBytes += int64(len(payload))
		if totalBytes > policy.MaxTotalBytes {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: fmt.Errorf("base64 media exceeds the aggregate limit of %d bytes", policy.MaxTotalBytes)}
		}
		digest := sha256.Sum256(payload)
		candidates = append(candidates, mediaCandidate{
			jsonPath:    container.jsonPath,
			contentType: contentType,
			payload:     payload,
			extension:   extension,
			hash:        hex.EncodeToString(digest[:]),
		})
	}
	return candidates, nil
}

func createRelayStorageObject(profile *model.StorageProfile, policy *model.StoragePolicy, candidate mediaCandidate, userID int, requestID string, taskID string, index int) (int64, string, error) {
	suffix, err := randomObjectSuffix()
	if err != nil {
		return 0, "", err
	}
	userHash := common.GenerateHMAC("storage-user:" + strconv.Itoa(userID))
	if len(userHash) > 24 {
		userHash = userHash[:24]
	}
	objectID := "obj_" + suffix
	objectKey := fmt.Sprintf("%s/volc-native/%s/user-%s/%s/%02d-%s.%s",
		strings.Trim(policy.ObjectPrefix, "/"),
		time.Now().UTC().Format("2006/01/02"),
		userHash,
		taskID,
		index+1,
		suffix,
		candidate.extension,
	)
	if len(objectKey) > 1024 {
		return 0, "", errors.New("temporary storage object key is too long")
	}
	object := &model.StorageObject{
		ObjectID:         objectID,
		OwnerUserID:      userID,
		RequestID:        requestID,
		TaskID:           taskID,
		StorageProfileID: profile.ID,
		StoragePolicyID:  policy.ID,
		Purpose:          model.StorageObjectPurposeRelayMediaTemp,
		ObjectKey:        objectKey,
		ContentType:      candidate.contentType,
		Size:             int64(len(candidate.payload)),
		SHA256:           candidate.hash,
		Status:           model.StorageObjectStatusUploading,
		DeleteAfter:      common.GetTimestamp() + policy.RetentionSeconds,
	}
	if err = model.CreateStorageObject(object); err != nil {
		return 0, "", err
	}
	return object.ID, objectKey, nil
}

func mediaExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return "jpg", true
	case "image/png":
		return "png", true
	case "image/webp":
		return "webp", true
	default:
		return "", false
	}
}
