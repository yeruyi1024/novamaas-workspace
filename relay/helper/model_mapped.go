package helper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func ModelMappedHelper(c *gin.Context, info *relaycommon.RelayInfo, request dto.Request) error {
	if info.ChannelMeta == nil {
		info.ChannelMeta = &relaycommon.ChannelMeta{}
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := common.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return fmt.Errorf("unmarshal_model_mapping_failed")
		}

		// 支持链式模型重定向，最终使用链尾的模型
		currentModel := info.OriginModelName
		visitedModels := map[string]bool{
			currentModel: true,
		}
		for {
			if mappedModel, exists := modelMap[currentModel]; exists && mappedModel != "" {
				// 模型重定向循环检测，避免无限循环
				if visitedModels[mappedModel] {
					if mappedModel == currentModel {
						if currentModel == info.OriginModelName {
							info.IsModelMapped = false
							return nil
						} else {
							info.IsModelMapped = true
							break
						}
					}
					return errors.New("model_mapping_contains_cycle")
				}
				visitedModels[mappedModel] = true
				currentModel = mappedModel
				info.IsModelMapped = true
			} else {
				break
			}
		}
		if info.IsModelMapped {
			info.UpstreamModelName = currentModel
		}
	}

	if request != nil {
		request.SetModelName(info.UpstreamModelName)
	}
	return nil
}

// ApplyModelMappingToJSONBody rewrites only the top-level model field. Raw
// native request bodies otherwise remain untouched, including provider-specific
// fields and number representations.
func ApplyModelMappingToJSONBody(info *relaycommon.RelayInfo, body []byte) ([]byte, error) {
	if info == nil || !info.IsModelMapped {
		return body, nil
	}
	if strings.TrimSpace(info.UpstreamModelName) == "" {
		return nil, errors.New("mapped upstream model is empty")
	}
	mappedBody, err := sjson.SetBytes(body, "model", info.UpstreamModelName)
	if err != nil {
		return nil, fmt.Errorf("set mapped model: %w", err)
	}
	return mappedBody, nil
}

// RestoreOriginalModelInJSONBody hides a mapped upstream model in a native
// response without adding a model field when the provider omitted it.
func RestoreOriginalModelInJSONBody(body []byte, originalModel string) ([]byte, error) {
	if strings.TrimSpace(originalModel) == "" || !gjson.GetBytes(body, "model").Exists() {
		return body, nil
	}
	restoredBody, err := sjson.SetBytes(body, "model", originalModel)
	if err != nil {
		return nil, fmt.Errorf("restore original model: %w", err)
	}
	return restoredBody, nil
}
