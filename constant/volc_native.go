package constant

import "strings"

// IsVolcNativeRequestPath matches only the registered native image/task routes.
func IsVolcNativeRequestPath(path string) bool {
	return path == "/api/v3/images/generations" ||
		path == "/api/v3/contents/generations/tasks" ||
		strings.HasPrefix(path, "/api/v3/contents/generations/tasks/")
}

// VolcNativeChannelMatchesPath isolates native channels in both directions.
// Empty paths are used by non-relay callers such as model discovery.
func VolcNativeChannelMatchesPath(channelType int, path string) bool {
	if path == "" {
		return true
	}
	nativePath := IsVolcNativeRequestPath(path)
	if nativePath || channelType == ChannelTypeVolcNative {
		return nativePath && channelType == ChannelTypeVolcNative
	}
	return true
}
