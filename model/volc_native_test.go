package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/stretchr/testify/require"
)

func TestVolcNativeTaskKeepsSubmissionCredentialPrivate(t *testing.T) {
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{ChannelType: constant.ChannelTypeVolcNative, ApiKey: "submission-key"}, TaskRelayInfo: &relaycommon.TaskRelayInfo{}}
	task := InitTask("61", info)
	require.Equal(t, "submission-key", task.PrivateData.Key)
	body, err := common.Marshal(task)
	require.NoError(t, err)
	require.NotContains(t, string(body), "submission-key")
}

func TestVolcNativeTaskPersistsOriginalAndUpstreamModels(t *testing.T) {
	info := &relaycommon.RelayInfo{
		OriginModelName: "public-seedance",
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:       constant.ChannelTypeVolcNative,
			UpstreamModelName: "doubao-seedance-2-0-260128",
			IsModelMapped:     true,
		},
		TaskRelayInfo: &relaycommon.TaskRelayInfo{},
	}

	task := InitTask("61", info)

	require.Equal(t, "public-seedance", task.Properties.OriginModelName)
	require.Equal(t, "doubao-seedance-2-0-260128", task.Properties.UpstreamModelName)
}

func TestVolcNativeChannelIsolationWithAndWithoutCache(t *testing.T) {
	setupChannelStatusTest(t)
	native := &Channel{Type: constant.ChannelTypeVolcNative}
	compatible := &Channel{Type: constant.ChannelTypeDoubaoVideo}
	require.NoError(t, DB.Create(native).Error)
	require.NoError(t, DB.Create(compatible).Error)
	channelSyncLock.Lock()
	oldChannels := channelsIDM
	channelsIDM = map[int]*Channel{native.Id: native, compatible.Id: compatible}
	channelSyncLock.Unlock()
	t.Cleanup(func() { channelSyncLock.Lock(); channelsIDM = oldChannels; channelSyncLock.Unlock() })
	for _, tc := range []struct {
		path string
		id   int
	}{
		{"/api/v3/contents/generations/tasks", native.Id},
		{"/api/v3/images/generations", native.Id},
		{"/v1/video/generations", compatible.Id},
	} {
		channelSyncLock.RLock()
		ids := filterChannelsByRequestPathAndModel([]int{native.Id, compatible.Id}, tc.path, "seedance")
		channelSyncLock.RUnlock()
		require.Equal(t, []int{tc.id}, ids)
		abilities := filterAbilitiesByRequestPathAndModel([]Ability{{ChannelId: native.Id}, {ChannelId: compatible.Id}}, tc.path, "seedance")
		require.Len(t, abilities, 1)
		require.Equal(t, tc.id, abilities[0].ChannelId)
	}
}
