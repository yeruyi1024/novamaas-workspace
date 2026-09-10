package storage

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestDecodeMediaCandidatesAcceptsMatchingImageDataURI(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52}
	container := mediaContainer{jsonPath: "content.0.image_url.url", dataURI: "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)}
	policy := model.DefaultRelayMediaStoragePolicy()

	candidates, err := decodeMediaCandidates([]mediaContainer{container}, policy)

	require.NoError(t, err)
	require.Len(t, candidates, 1)
	assert.Equal(t, "image/png", candidates[0].contentType)
	assert.Equal(t, "png", candidates[0].extension)
	assert.Equal(t, png, candidates[0].payload)
}

func TestDecodeMediaCandidatesAcceptsCaseInsensitiveDataURIMetadata(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	container := mediaContainer{jsonPath: "content.0.image_url.url", dataURI: "DATA:IMAGE/PNG;BASE64," + base64.StdEncoding.EncodeToString(png)}

	candidates, err := decodeMediaCandidates([]mediaContainer{container}, model.DefaultRelayMediaStoragePolicy())

	require.NoError(t, err)
	require.Len(t, candidates, 1)
	assert.Equal(t, "image/png", candidates[0].contentType)
}

func TestDecodeMediaCandidatesRejectsDeclaredTypeMismatch(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	container := mediaContainer{jsonPath: "content.0.image_url.url", dataURI: "data:image/webp;base64," + base64.StdEncoding.EncodeToString(png)}

	_, err := decodeMediaCandidates([]mediaContainer{container}, model.DefaultRelayMediaStoragePolicy())

	var requestErr *RequestError
	require.ErrorAs(t, err, &requestErr)
	assert.Equal(t, http.StatusBadRequest, requestErr.StatusCode)
	assert.Contains(t, requestErr.Error(), "type mismatch")
}

func TestFindVideoTaskDataURIContainersFindsImagesAndVideos(t *testing.T) {
	request := []byte(`{"seed":9007199254740993,"content":[{"image_url":{"url":"https://example.com/reference.webp"}},{"image_url":{"url":"data:image/webp;base64,UklGRg=="}},{"video_url":{"url":"data:video/mp4;base64,AAAA"}}]}`)

	containers, err := findVideoTaskDataURIContainers(request)

	require.NoError(t, err)
	require.Len(t, containers, 2)
	assert.Equal(t, "content.1.image_url.url", containers[0].jsonPath)
	assert.Equal(t, "image_url", containers[0].mediaType)
	assert.Equal(t, "data:image/webp;base64,UklGRg==", containers[0].dataURI)
	assert.Equal(t, "content.2.video_url.url", containers[1].jsonPath)
	assert.Equal(t, "video_url", containers[1].mediaType)
	assert.Equal(t, "data:video/mp4;base64,AAAA", containers[1].dataURI)
}

func TestDecodeMediaCandidatesAcceptsMatchingVideoDataURI(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		extension   string
		payload     []byte
	}{
		{name: "MP4", contentType: "video/mp4", extension: "mp4", payload: []byte{0x00, 0x00, 0x00, 0x0c, 'f', 't', 'y', 'p', 'm', 'p', '4', '2'}},
		{name: "WebM", contentType: "video/webm", extension: "webm", payload: []byte{0x1a, 0x45, 0xdf, 0xa3}},
		{name: "QuickTime", contentType: "video/quicktime", extension: "mov", payload: []byte{0x00, 0x00, 0x00, 0x0c, 'f', 't', 'y', 'p', 'q', 't', ' ', ' '}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			container := mediaContainer{
				jsonPath:  "content.0.video_url.url",
				mediaType: "video_url",
				dataURI:   "data:" + test.contentType + ";base64," + base64.StdEncoding.EncodeToString(test.payload),
			}

			candidates, err := decodeMediaCandidates([]mediaContainer{container}, model.DefaultRelayMediaStoragePolicy())

			require.NoError(t, err)
			require.Len(t, candidates, 1)
			assert.Equal(t, test.contentType, candidates[0].contentType)
			assert.Equal(t, test.extension, candidates[0].extension)
			assert.Equal(t, test.payload, candidates[0].payload)
		})
	}
}

func TestDecodeMediaCandidatesRejectsMediaFieldTypeMismatch(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	container := mediaContainer{
		jsonPath:  "content.0.video_url.url",
		mediaType: "video_url",
		dataURI:   "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
	}

	_, err := decodeMediaCandidates([]mediaContainer{container}, model.DefaultRelayMediaStoragePolicy())

	var requestErr *RequestError
	require.ErrorAs(t, err, &requestErr)
	assert.Equal(t, http.StatusBadRequest, requestErr.StatusCode)
	assert.Contains(t, requestErr.Error(), "video_url must contain video media")
}

func TestDecodeMediaCandidatesEnforcesAggregateLimitBeforeUpload(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	dataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	policy := model.DefaultRelayMediaStoragePolicy()
	policy.MaxFileBytes = int64(len(png))
	policy.MaxTotalBytes = int64(len(png))

	_, err := decodeMediaCandidates([]mediaContainer{{dataURI: dataURI}, {dataURI: dataURI}}, policy)

	var requestErr *RequestError
	require.ErrorAs(t, err, &requestErr)
	assert.Equal(t, http.StatusBadRequest, requestErr.StatusCode)
	assert.Contains(t, requestErr.Error(), "aggregate limit")
}

func TestReplaceMaterializedMediaURLPreservesLargeIntegerLiteral(t *testing.T) {
	body := []byte(`{"seed":9007199254740993,"content":[{"image_url":{"url":"data:image/webp;base64,UklGRg=="}}]}`)
	containers, err := findVideoTaskDataURIContainers(body)
	require.NoError(t, err)
	require.Len(t, containers, 1)

	replaced, err := replaceMaterializedMediaURL(body, containers[0].jsonPath, "https://example.com/image.webp")
	require.NoError(t, err)

	assert.Contains(t, string(replaced), `"seed":9007199254740993`)
	assert.Equal(t, "https://example.com/image.webp", gjson.GetBytes(replaced, "content.0.image_url.url").String())
}
