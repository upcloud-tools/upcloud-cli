package storage

import (
	"encoding/json"
	"testing"

	"github.com/UpCloudLtd/upcloud-cli/v3/internal/commands"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/config"
	smock "github.com/UpCloudLtd/upcloud-cli/v3/internal/mock"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/mockexecute"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListStorages(t *testing.T) {
	Storage1 := upcloud.Storage{
		UUID:   UUID1,
		Title:  Title1,
		Access: "private",
		State:  "maintenance",
		Type:   "backup",
		Zone:   "fi-hel1",
		Size:   40,
		Tier:   "maxiops",
	}
	Storage2 := upcloud.Storage{
		UUID:      UUID2,
		Title:     Title2,
		Encrypted: upcloud.FromBool(true),
		Access:    "private",
		State:     "online",
		Type:      "normal",
		Zone:      "fi-hel1",
		Size:      40,
		Tier:      "maxiops",
	}
	Storage3 := upcloud.Storage{
		UUID:   UUID3,
		Title:  Title3,
		Access: "public",
		State:  "online",
		Type:   "normal",
		Zone:   "fi-hel1",
		Size:   10,
		Tier:   "maxiops",
	}
	Storage4 := Storage1
	Storage4.Title = "mock-storage-title4"
	Storage4.Type = upcloud.StorageTypeCDROM
	Storage5 := Storage1
	Storage5.Title = "mock-storage-title5"
	Storage5.Type = upcloud.StorageTypeTemplate
	Storage6 := Storage3
	Storage6.Title = "mock-storage-title6"
	Storage6.Type = upcloud.StorageTypeBackup

	allStorages := upcloud.Storages{Storages: []upcloud.Storage{Storage1, Storage2, Storage3, Storage4, Storage5, Storage6}}
	storageTitles := make([]string, 6)
	for i, storage := range allStorages.Storages {
		storageTitles[i] = storage.Title
	}

	for _, test := range []struct {
		name              string
		private           bool
		public            bool
		args              []string
		outputContains    []string
		outputNotContains []string
	}{
		{
			name:           "List storages",
			args:           []string{"--all"},
			outputContains: storageTitles,
		},
		{
			name:              "List public storages",
			args:              []string{"--public"},
			outputContains:    []string{Storage3.Title, Storage6.Title},
			outputNotContains: []string{Storage1.Title, Storage2.Title, Storage4.Title, Storage5.Title},
		},
		{
			name:              "List private by default",
			args:              []string{},
			outputContains:    []string{Storage1.Title, Storage2.Title, Storage4.Title, Storage5.Title},
			outputNotContains: []string{Storage3.Title, Storage6.Title},
		},
		{
			name:              "List cdrom",
			args:              []string{"--cdrom"},
			outputContains:    []string{Storage4.Title},
			outputNotContains: []string{Storage1.Title, Storage2.Title, Storage3.Title, Storage5.Title, Storage6.Title},
		},
		{
			name:              "List public backup",
			args:              []string{"--public", "--backup"},
			outputContains:    []string{Storage6.Title},
			outputNotContains: []string{Storage1.Title, Storage2.Title, Storage3.Title, Storage4.Title, Storage5.Title},
		},
		{
			name:              "List public template",
			args:              []string{"--public", "--template"},
			outputNotContains: []string{Storage1.Title, Storage2.Title, Storage3.Title, Storage4.Title, Storage5.Title, Storage6.Title},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			CachedStorages = nil
			conf := config.New()
			mService := new(smock.Service)

			mService.On("GetStorages", mock.Anything).Return(&allStorages, nil)

			c := commands.BuildCommand(ListCommand(), nil, config.New())

			c.Cobra().SetArgs(test.args)
			output, err := mockexecute.MockExecute(c, mService, conf)

			assert.NoError(t, err)
			mService.AssertNumberOfCalls(t, "GetStorages", 1)

			for _, contains := range test.outputContains {
				assert.Contains(t, output, contains)
			}

			for _, notContains := range test.outputNotContains {
				assert.NotContains(t, output, notContains)
			}
		})
	}
}

func TestListStoragesAttachedServers(t *testing.T) {
	text.DisableColors()
	attached := upcloud.Storage{
		UUID:   UUID1,
		Title:  Title1,
		Access: "private",
		State:  "online",
		Type:   "normal",
		Zone:   "fi-hel1",
		Size:   40,
		Tier:   "maxiops",
	}
	detached := upcloud.Storage{
		UUID:   UUID2,
		Title:  Title2,
		Access: "private",
		State:  "online",
		Type:   "normal",
		Zone:   "fi-hel1",
		Size:   10,
		Tier:   "maxiops",
	}
	serverUUID := "009d64ef-31d1-4684-a26b-c86c955cbf46"
	storages := upcloud.Storages{Storages: []upcloud.Storage{attached, detached}}

	for _, test := range []struct {
		name         string
		args         []string
		jsonOutput   bool
		detailsCalls int
		contains     []string
		notContains  []string
		checkJSON    bool
	}{
		{
			name:         "human output with attached servers",
			args:         []string{"--show-attached-servers"},
			detailsCalls: 2,
			contains:     []string{"Attached to", serverUUID, "None"},
		},
		{
			name:         "human output without flag",
			args:         []string{},
			detailsCalls: 0,
			notContains:  []string{"Attached to", serverUUID},
		},
		{
			name:         "JSON output with attached servers",
			args:         []string{"--show-attached-servers"},
			jsonOutput:   true,
			detailsCalls: 2,
			checkJSON:    true,
		},
		{
			name:         "JSON output without flag",
			args:         []string{},
			jsonOutput:   true,
			detailsCalls: 0,
			notContains:  []string{`"servers"`},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			CachedStorages = nil
			mService := new(smock.Service)
			mService.On("GetStorages", mock.Anything).Return(&storages, nil)
			mService.On("GetStorageDetails", &request.GetStorageDetailsRequest{UUID: UUID1}).Return(&upcloud.StorageDetails{
				Storage:     attached,
				ServerUUIDs: upcloud.ServerUUIDSlice{serverUUID},
			}, nil)
			mService.On("GetStorageDetails", &request.GetStorageDetailsRequest{UUID: UUID2}).Return(&upcloud.StorageDetails{
				Storage:     detached,
				ServerUUIDs: upcloud.ServerUUIDSlice{},
			}, nil)

			conf := config.New()
			if test.jsonOutput {
				conf.Viper().Set(config.KeyOutput, config.ValueOutputJSON)
			}

			c := commands.BuildCommand(ListCommand(), nil, config.New())
			c.Cobra().SetArgs(test.args)
			output, err := mockexecute.MockExecute(c, mService, conf)
			assert.NoError(t, err)
			mService.AssertNumberOfCalls(t, "GetStorageDetails", test.detailsCalls)

			for _, contains := range test.contains {
				assert.Contains(t, output, contains)
			}
			for _, notContains := range test.notContains {
				assert.NotContains(t, output, notContains)
			}

			if test.checkJSON {
				var parsed struct {
					Storages []map[string]any `json:"storages"`
				}
				require.NoError(t, json.Unmarshal([]byte(output), &parsed))
				require.Len(t, parsed.Storages, 2)
				for _, s := range parsed.Storages {
					// servers uses the same embedded shape as the storage details API response
					serversBlock, ok := s["servers"].(map[string]any)
					require.True(t, ok, "storage %v should include servers", s["uuid"])
					servers, _ := serversBlock["server"].([]any)
					if s["uuid"] == UUID1 {
						assert.Equal(t, []any{serverUUID}, servers)
					} else {
						assert.Empty(t, servers)
					}
				}
			}
		})
	}
}
