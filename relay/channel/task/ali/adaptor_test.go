package ali

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRelayInfo() *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		ChannelMeta:   &relaycommon.ChannelMeta{},
		TaskRelayInfo: &relaycommon.TaskRelayInfo{},
	}
}

func TestConvertToAliRequestWan27I2VBuildsMediaFromImage(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:    "wan2.7-i2v",
		Prompt:   "animate the first frame",
		Image:    "https://example.com/first.png",
		Size:     "720p",
		Duration: 10,
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, "wan2.7-i2v", aliReq.Model)
	require.Equal(t, "720P", aliReq.Parameters.Resolution)
	assert.Equal(t, 10, aliReq.Parameters.Duration)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
	}, aliReq.Input.Media)
	require.Empty(t, aliReq.Input.ImgURL)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"media"`)
	require.NotContains(t, string(body), `"img_url"`)
}

func TestConvertToAliRequestWan27I2VBuildsFirstAndLastFrameFromImages(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "interpolate between frames",
		Images: []string{
			"https://example.com/first.png",
			"https://example.com/last.png",
		},
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
		{Type: "last_frame", URL: "https://example.com/last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VPrefersImageBeforeImagesAndInputReference(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "wan2.7-i2v",
		Prompt:         "use the direct image",
		Image:          " https://example.com/direct.png ",
		Images:         []string{"https://example.com/images-first.png", " https://example.com/images-last.png "},
		InputReference: "https://example.com/input-reference.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/direct.png"},
		{Type: "last_frame", URL: "https://example.com/images-last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VFallsBackToFirstNonEmptyImage(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "skip blank images",
		Image:  " ",
		Images: []string{
			" ",
			" https://example.com/first.png ",
			" https://example.com/last.png ",
		},
		InputReference: "https://example.com/input-reference.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_frame", URL: "https://example.com/first.png"},
		{Type: "last_frame", URL: "https://example.com/last.png"},
	}, aliReq.Input.Media)
}

func TestConvertToAliRequestWan27I2VKeepsExplicitMetadataMedia(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "wan2.7-i2v",
		Prompt:         "continue the clip",
		Image:          "https://example.com/direct.png",
		Images:         []string{"https://example.com/images-first.png", "https://example.com/images-last.png"},
		InputReference: "https://example.com/input-reference.png",
		Metadata: map[string]interface{}{
			"input": map[string]interface{}{
				"media": []interface{}{
					map[string]interface{}{
						"type": "first_clip",
						"url":  "https://example.com/input.mp4",
					},
				},
			},
		},
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, []AliVideoMedia{
		{Type: "first_clip", URL: "https://example.com/input.mp4"},
	}, aliReq.Input.Media)
	require.Empty(t, aliReq.Input.ImgURL)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"media"`)
	require.NotContains(t, string(body), `"img_url"`)
}

func TestConvertToAliRequestWan27I2VRequiresMedia(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.7-i2v",
		Prompt: "animate without a frame",
	}

	_, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "requires image"))
}

func TestConvertToAliRequestWan25I2VKeepsLegacyImgURL(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "wan2.5-i2v-preview",
		Prompt: "animate the first frame",
		Image:  "https://example.com/first.png",
	}

	aliReq, err := adaptor.convertToAliRequest(testRelayInfo(), req)

	require.NoError(t, err)
	require.Equal(t, "https://example.com/first.png", aliReq.Input.ImgURL)
	require.Empty(t, aliReq.Input.Media)

	body, err := common.Marshal(aliReq)
	require.NoError(t, err)
	require.Contains(t, string(body), `"img_url"`)
	require.NotContains(t, string(body), `"media"`)
}

func TestBuildRequestBodyRecordsAliUpstreamSnapshot(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("task_request", relaycommon.TaskSubmitReq{
		Model:  "wan2.5-i2v-preview",
		Prompt: "animate the first frame",
		Image:  "https://example.com/first.png",
	})
	info := testRelayInfo()
	info.OriginModelName = "wan2.5-i2v-preview"

	body, err := (&TaskAdaptor{}).BuildRequestBody(c, info)
	require.NoError(t, err)
	upstreamBody, err := io.ReadAll(body)
	require.NoError(t, err)

	assert.JSONEq(t, string(upstreamBody), common.GetContextKeyString(c, constant.ContextKeyVideoTaskUpstreamRequestBody))
}

func TestTaskAdaptorParseTaskResultAcceptsWan3FractionalUsage(t *testing.T) {
	adaptor := &TaskAdaptor{}
	responseBody := []byte(`{
		"request_id": "request-id",
		"output": {
			"task_id": "upstream-task-id",
			"task_status": "SUCCEEDED",
			"video_url": "https://example.com/video.mp4"
		},
		"usage": {
			"video_count": 1,
			"duration": 5.0,
			"SR": 720,
			"output_video_duration": 5.0,
			"input_video_duration": 0.0,
			"fps": 30,
			"ratio": "16:9"
		}
	}`)

	result, err := adaptor.ParseTaskResult(responseBody)

	require.NoError(t, err)
	assert.Equal(t, model.TaskStatusSuccess, result.Status)
	assert.Equal(t, "https://example.com/video.mp4", result.Url)
}

func TestWan3RequestCompatibility(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		parameters    string
		input         string
		media         bool
		seconds       float64
		upstreamModel string
	}{
		{
			name:       "mapped wan3 model keeps flat parameters",
			body:       `{"model":"public-video","prompt":"animate","image":"https://example.com/image.png","audio":"false","resolution":"720P","ratio":"16:9"}`,
			parameters: `{"audio":false,"resolution":"720P","ratio":"16:9","duration":5,"prompt_extend":true}`,
			media:      true, seconds: 5, upstreamModel: "wan3.0-video",
		},
		{
			name:       "flat string audio and legacy image",
			body:       `{"model":"wan3.0-video","prompt":"animate","image":"https://example.com/image.png","audio":"false","ratio":"16:9","resolution":"720P","duration":10}`,
			parameters: `{"audio":false,"ratio":"16:9","resolution":"720P","duration":10,"prompt_extend":true}`,
			media:      true, seconds: 10,
		},
		{
			name:       "metadata size tier and boolean string",
			body:       `{"model":"wan3.0-video","prompt":"animate","metadata":{"parameters":{"size":"1080p","audio":"false"}}}`,
			parameters: `{"audio":false,"resolution":"1080P","duration":5,"prompt_extend":true}`,
			seconds:    5,
		},
		{
			name:       "existing nested metadata preserves false and zero",
			body:       `{"model":"wan3.0-video","prompt":"animate","metadata":{"parameters":{"audio":false,"ratio":"9:16","resolution":"480P","prompt_extend":false,"watermark":false,"seed":0,"duration":8}}}`,
			parameters: `{"audio":false,"ratio":"9:16","resolution":"480P","duration":8,"prompt_extend":false,"watermark":false,"seed":0}`,
			seconds:    8,
		},
		{
			name:       "official input and parameters",
			body:       `{"model":"wan3.0-video-prime","input":{"prompt":"animate","media":[{"type":"reference_image","url":"https://example.com/image.png"}]},"parameters":{"audio":"false","ratio":"adaptive","resolution":"720P","duration":12}}`,
			parameters: `{"audio":false,"ratio":"adaptive","resolution":"720P","duration":12,"prompt_extend":true}`,
			input:      `{"prompt":"animate","media":[{"type":"reference_image","url":"https://example.com/image.png"}]}`,
			media:      true, seconds: 12,
		},
		{
			name:       "official references preserve order and override legacy images",
			body:       `{"model":"wan3.0-video","image":"https://example.com/unused.png","images":["https://example.com/unused-first.png","https://example.com/unused-last.png"],"input":{"prompt":"reference image 2 and video 1 with audio 1","media":[{"type":"reference_image","url":"https://example.com/one.png"},{"type":"reference_video","url":"https://example.com/video.mp4"},{"type":"reference_image","url":"https://example.com/two.png"},{"type":"reference_audio","url":"https://example.com/audio.mp3"}]}}`,
			parameters: `{"resolution":"720P","duration":5,"prompt_extend":true}`,
			input:      `{"prompt":"reference image 2 and video 1 with audio 1","media":[{"type":"reference_image","url":"https://example.com/one.png"},{"type":"reference_video","url":"https://example.com/video.mp4"},{"type":"reference_image","url":"https://example.com/two.png"},{"type":"reference_audio","url":"https://example.com/audio.mp3"}]}`,
			media:      true, seconds: 5,
		},
		{
			name:       "official first and last frames preserve media URLs",
			body:       `{"model":"wan3.0-video","input":{"prompt":"transition","media":[{"type":"first_frame","url":"data:image/png;base64,aGVsbG8="},{"type":"last_frame","url":"oss://dashscope-instant/last.png"}]}}`,
			parameters: `{"resolution":"720P","duration":5,"prompt_extend":true}`,
			input:      `{"prompt":"transition","media":[{"type":"first_frame","url":"data:image/png;base64,aGVsbG8="},{"type":"last_frame","url":"oss://dashscope-instant/last.png"}]}`,
			media:      true, seconds: 5,
		},
		{
			name:       "official file input without parameters",
			body:       `{"model":"wan3.0-video","input":{"media":[{"type":"file","url":"https://example.com/document.pdf"}]}}`,
			parameters: `{"resolution":"720P","duration":5,"prompt_extend":true}`,
			input:      `{"media":[{"type":"file","url":"https://example.com/document.pdf"}]}`,
			media:      true, seconds: 5,
		},
		{
			name:       "metadata link input overrides official media",
			body:       `{"model":"wan3.0-video","prompt":"summarize","input":{"media":[{"type":"first_frame","url":"https://example.com/unused.png"}]},"metadata":{"input":{"media":[{"type":"link","url":"https://example.com/article"},{"type":"reference_audio","url":"https://example.com/audio.mp3"}]}}}`,
			parameters: `{"resolution":"720P","duration":5,"prompt_extend":true}`,
			input:      `{"prompt":"summarize","media":[{"type":"link","url":"https://example.com/article"},{"type":"reference_audio","url":"https://example.com/audio.mp3"}]}`,
			media:      true, seconds: 5,
		},
		{
			name:       "metadata wins per field without dropping flat ratio",
			body:       `{"model":"wan3.0-video","prompt":"animate","audio":true,"ratio":"16:9","parameters":{"audio":true,"resolution":"720P","duration":10},"metadata":{"parameters":{"audio":"false","duration":6}}}`,
			parameters: `{"audio":false,"ratio":"16:9","resolution":"720P","duration":6,"prompt_extend":true}`,
			seconds:    6,
		},
		{
			name:       "legacy pixel size maps to official fields",
			body:       `{"model":"wan3.0-video","prompt":"animate","size":"1280*720"}`,
			parameters: `{"ratio":"16:9","resolution":"720P","duration":5,"prompt_extend":true}`,
			seconds:    5,
		},
		{
			name:       "explicit resolution overrides legacy pixel size",
			body:       `{"model":"wan3.0-video","prompt":"animate","size":"1280x720","parameters":{"resolution":"1080P","ratio":"9:16"}}`,
			parameters: `{"ratio":"9:16","resolution":"1080P","duration":5,"prompt_extend":true}`,
			seconds:    5,
		},
		{
			name:       "official media without prompt",
			body:       `{"model":"wan3.0-video","input":{"media":[{"type":"first_frame","url":"https://example.com/image.png"}]}}`,
			parameters: `{"resolution":"720P","duration":5,"prompt_extend":true}`,
			media:      true, seconds: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			t.Cleanup(func() { common.CleanupBodyStorage(c) })
			info := testRelayInfo()
			adaptor := &TaskAdaptor{}
			if tt.upstreamModel != "" {
				c.Set("model_mapping", `{"public-video":"wan3.0-video"}`)
			}
			require.Nil(t, adaptor.ValidateRequestAndSetAction(c, info))
			if tt.upstreamModel != "" {
				info.IsModelMapped = true
				info.UpstreamModelName = tt.upstreamModel
			}
			body, err := adaptor.BuildRequestBody(c, info)
			require.NoError(t, err)
			data, err := io.ReadAll(body)
			require.NoError(t, err)
			var upstream struct {
				Model      string         `json:"model"`
				Input      AliVideoInput  `json:"input"`
				Parameters map[string]any `json:"parameters"`
			}
			require.NoError(t, common.Unmarshal(data, &upstream))
			if tt.upstreamModel != "" {
				assert.Equal(t, tt.upstreamModel, upstream.Model)
			}
			parameters, err := common.Marshal(upstream.Parameters)
			require.NoError(t, err)
			assert.JSONEq(t, tt.parameters, string(parameters))
			if tt.input != "" {
				input, err := common.Marshal(upstream.Input)
				require.NoError(t, err)
				assert.JSONEq(t, tt.input, string(input))
			}
			assert.Equal(t, tt.media, len(upstream.Input.Media) > 0)
			assert.Empty(t, upstream.Input.ImgURL)
			assert.Equal(t, tt.seconds, adaptor.EstimateBilling(c, info)["seconds"])
		})
	}
}

func TestWan3RejectsInvalidParametersBeforeBilling(t *testing.T) {
	for _, parameters := range []string{
		`{"audio":"invalid"}`, `{"ratio":"2:1"}`, `{"resolution":"4K"}`,
		`{"duration":-1}`, `{"duration":0}`, `{"duration":31}`, `{"duration":3601}`,
		`null`,
	} {
		t.Run(parameters, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", strings.NewReader(`{"model":"wan3.0-video","prompt":"animate","metadata":{"parameters":`+parameters+`}}`))
			c.Request.Header.Set("Content-Type", "application/json")
			t.Cleanup(func() { common.CleanupBodyStorage(c) })
			err := (&TaskAdaptor{}).ValidateRequestAndSetAction(c, testRelayInfo())
			require.NotNil(t, err)
			assert.Equal(t, http.StatusBadRequest, err.StatusCode)
			_, stored := c.Get("task_request")
			assert.False(t, stored)
		})
	}
}

func TestWan3ChannelParameterOverrides(t *testing.T) {
	for _, tt := range []struct {
		name, body, condition, path string
		upstreamModel               string
		value                       any
		resolution                  string
		duration                    float64
		invalid                     bool
	}{
		{name: "official media with configured size", body: `{"model":"wan3.0-video-480p","input":{"prompt":"animate","media":[{"type":"reference_image","url":"https://example.com/a.png"},{"type":"reference_audio","url":"https://example.com/a.mp3"}]},"parameters":{"audio":false}}`, condition: "wan3.0-video-480p", path: "size", value: "480p", resolution: "480P", duration: 5},
		{name: "flat request", body: `{"model":"wan3.0-video-480p","prompt":"animate","audio":"false"}`, condition: "wan3.0-video-480p", path: "size", value: "480p", resolution: "480P", duration: 5},
		{name: "legacy metadata request", body: `{"model":"wan3.0-video-480p","prompt":"animate","metadata":{"parameters":{"audio":false}}}`, condition: "wan3.0-video-480p", path: "size", value: "480p", resolution: "480P", duration: 5},
		{name: "screenshot typo does not match", body: `{"model":"wan3.0-video-480p","prompt":"animate"}`, condition: "wan3-video-480p", path: "size", value: "480p", resolution: "720P", duration: 5},
		{name: "configured arbitrary public alias", body: `{"model":"customer-video","prompt":"animate"}`, condition: "customer-video", path: "size", value: "1080p", resolution: "1080P", duration: 5},
		{name: "explicit resolution still takes priority over size", body: `{"model":"wan3.0-video-480p","prompt":"animate","parameters":{"resolution":"1080P"}}`, condition: "wan3.0-video-480p", path: "size", value: "480p", resolution: "1080P", duration: 5},
		{name: "nested resolution override", body: `{"model":"wan3.0-video-480p","prompt":"animate","parameters":{"resolution":"1080P"}}`, condition: "wan3.0-video-480p", path: "parameters.resolution", value: "480P", resolution: "480P", duration: 5},
		{name: "duration used by billing and upstream", body: `{"model":"wan3.0-video-480p","prompt":"animate"}`, condition: "wan3.0-video-480p", path: "parameters.duration", value: 8, resolution: "720P", duration: 8},
		{name: "invalid override rejected before billing", body: `{"model":"wan3.0-video-480p","prompt":"animate"}`, condition: "wan3.0-video-480p", path: "parameters.duration", value: 1000000, invalid: true},
		{name: "legacy Ali model stays outside new override path", body: `{"model":"wan2.5-i2v-preview","prompt":"animate"}`, condition: "wan2.5-i2v-preview", path: "size", value: "480p", resolution: "1080P", duration: 5, upstreamModel: "wan2.5-i2v-preview"},
		{name: "model changes require model mapping", body: `{"model":"wan3.0-video-480p","prompt":"animate"}`, condition: "wan3.0-video-480p", path: "model", value: "different-model", invalid: true},
		{name: "invalid size rejected", body: `{"model":"wan3.0-video-480p","prompt":"animate"}`, condition: "wan3.0-video-480p", path: "size", value: "invalid", invalid: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("model_mapping", `{"wan3.0-video-480p":"wan3.0-video","customer-video":"wan3.0-video"}`)
			info := testRelayInfo()
			var original relaycommon.TaskSubmitReq
			require.NoError(t, common.Unmarshal([]byte(tt.body), &original))
			info.OriginModelName = original.Model
			info.UpstreamModelName = original.Model
			info.ParamOverride = map[string]any{"operations": []any{map[string]any{
				"path": tt.path, "mode": "set", "value": tt.value, "logic": "AND",
				"conditions": []any{map[string]any{"path": "original_model", "mode": "full", "value": tt.condition}},
			}}}
			adaptor := &TaskAdaptor{}
			taskErr := adaptor.ValidateRequestAndSetAction(c, info)
			storage, err := common.GetBodyStorage(c)
			require.NoError(t, err)
			t.Cleanup(func() { common.CleanupBodyStorage(c) })
			raw, err := storage.Bytes()
			require.NoError(t, err)
			assert.Equal(t, tt.body, string(raw))
			if tt.invalid {
				require.NotNil(t, taskErr)
				assert.Equal(t, http.StatusBadRequest, taskErr.StatusCode)
				assert.Nil(t, adaptor.wan3Request)
				return
			}
			require.Nil(t, taskErr)
			require.NoError(t, helper.ModelMappedHelper(c, info, nil))
			assert.Equal(t, original.Model, info.OriginModelName, "channel overrides must preserve the pricing model")
			assert.Equal(t, tt.duration, adaptor.EstimateBilling(c, info)["seconds"])
			reader, err := adaptor.BuildRequestBody(c, info)
			require.NoError(t, err)
			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			var request wan3VideoRequest
			require.NoError(t, common.Unmarshal(body, &request))
			expectedModel := tt.upstreamModel
			if expectedModel == "" {
				expectedModel = "wan3.0-video"
			}
			assert.Equal(t, expectedModel, request.Model)
			assert.Equal(t, tt.resolution, request.Parameters.Resolution)
			assert.Equal(t, int(tt.duration), *request.Parameters.Duration)
			if strings.Contains(tt.body, `"audio"`) {
				require.NotNil(t, request.Parameters.Audio)
				assert.False(t, bool(*request.Parameters.Audio))
			}
			if original.Model == "wan3.0-video-480p" && strings.Contains(tt.body, `"media"`) {
				assert.Equal(t, []AliVideoMedia{{Type: "reference_image", URL: "https://example.com/a.png"}, {Type: "reference_audio", URL: "https://example.com/a.mp3"}}, request.Input.Media)
			}
		})
	}
}
