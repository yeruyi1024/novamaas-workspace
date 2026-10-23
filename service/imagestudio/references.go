package imagestudio

import (
	"bytes"
	"context"
	"errors"
	"image"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	storageService "github.com/QuantumNous/new-api/service/storage"
)

const (
	MaxReferenceImages     = 8
	MaxReferenceImageBytes = 20 << 20
)

type ReferenceView struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	ExpiresAt int64  `json:"expires_at"`
	CreatedAt int64  `json:"created_at"`
}

type ReferenceList struct {
	Items []ReferenceView `json:"items"`
	Total int64           `json:"total"`
}

func ListReferences(ctx context.Context, userID int, page *common.PageInfo) (*ReferenceList, error) {
	objects, total, err := model.ListImageStudioReferenceObjects(ctx, userID, page)
	if err != nil {
		return nil, err
	}
	list := &ReferenceList{Items: make([]ReferenceView, 0, len(objects)), Total: total}
	for _, object := range objects {
		value, err := storageService.PresignImageStudioReference(ctx, object.ObjectID, userID)
		if err != nil {
			return nil, err
		}
		list.Items = append(list.Items, ReferenceView{ID: object.ObjectID, URL: value, ExpiresAt: object.DeleteAfter, CreatedAt: object.CreatedAt})
	}
	return list, nil
}

func ImportReference(ctx context.Context, userID int, source string) (*ReferenceView, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	data, err := readImage(ctx, dto.ImageData{Url: source}, MaxReferenceImageBytes)
	if err != nil {
		return nil, &RequestError{http.StatusBadRequest, "invalid_reference_image", errors.New("Could not read the reference image. Upload the file instead.")}
	}
	return UploadReference(ctx, userID, data)
}

func UploadReference(ctx context.Context, userID int, data []byte) (*ReferenceView, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if len(data) == 0 || len(data) > MaxReferenceImageBytes {
		return nil, &RequestError{http.StatusBadRequest, "invalid_reference_image", errors.New("Reference images must be PNG, JPEG or WebP and no larger than 20 MB.")}
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	contentType := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "webp": "image/webp"}[format]
	if err != nil || contentType == "" || config.Width <= 0 || config.Height <= 0 {
		return nil, &RequestError{http.StatusBadRequest, "invalid_reference_image", errors.New("Reference images must be PNG, JPEG or WebP and no larger than 20 MB.")}
	}
	object, err := storageService.UploadImageStudioReference(ctx, userID, contentType, data)
	if err != nil {
		status := http.StatusBadGateway
		var storageErr *storageService.RequestError
		if errors.As(err, &storageErr) {
			status = storageErr.StatusCode
		}
		return nil, &RequestError{status, "reference_upload_failed", err}
	}
	value, err := storageService.PresignImageStudioReference(ctx, object.ObjectID, userID)
	if err != nil {
		_ = storageService.DeleteAssetObject(object.ID, "reference URL signing failed")
		return nil, err
	}
	return &ReferenceView{ID: object.ObjectID, URL: value, ExpiresAt: object.DeleteAfter, CreatedAt: object.CreatedAt}, nil
}
