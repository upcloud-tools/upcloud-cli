package nodegroup

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/UpCloudLtd/upcloud-cli/v3/internal/commands"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/config"
	smock "github.com/UpCloudLtd/upcloud-cli/v3/internal/mock"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/mockexecute"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaleKubernetesNodeGroup(t *testing.T) {
	clusterUUID := "28c80353-98fd-4221-85e0-82d7603756ba"
	path := "/1.3/kubernetes/" + clusterUUID + "/node-groups/my-node-group"
	nodeGroup := func(count int) string {
		return `{"name": "my-node-group", "count": ` + strconv.Itoa(count) + `, "plan": "dev-md", "state": "running"}`
	}

	for _, test := range []struct {
		name      string
		args      []string
		responses map[string]fakeResponse
		body      string
		errorMsg  string
	}{
		{
			name:     "no args",
			args:     []string{clusterUUID},
			errorMsg: `required flag(s) "name", "count" not set`,
		},
		{
			name: "scale to positive count",
			args: []string{
				clusterUUID,
				"--name", "my-node-group",
				"--count", "6",
			},
			responses: map[string]fakeResponse{"PATCH " + path: {status: http.StatusOK, body: nodeGroup(6)}},
			body:      `{"count": 6}`,
		},
		{
			name: "scale to zero",
			args: []string{
				clusterUUID,
				"--name", "my-node-group",
				"--count", "0",
			},
			responses: map[string]fakeResponse{"PATCH " + path: {status: http.StatusOK, body: nodeGroup(0)}},
			body:      `{"count": 0}`,
		},
		{
			name: "API error is reported",
			args: []string{
				clusterUUID,
				"--name", "my-node-group",
				"--count", "0",
			},
			responses: map[string]fakeResponse{"PATCH " + path: {
				status:      http.StatusUnprocessableEntity,
				contentType: "application/problem+json",
				body:        `{"type": "https://developers.upcloud.com/1.3/errors#ERROR_INVALID_REQUEST", "title": "Invalid request", "status": 422}`,
			}},
			body:     `{"count": 0}`,
			errorMsg: "Invalid request (type=https://developers.upcloud.com/1.3/errors#ERROR_INVALID_REQUEST, status=422)",
		},
		{
			name: "invalid UUID",
			args: []string{
				"not-a-uuid",
				"--name", "my-node-group",
				"--count", "0",
			},
			errorMsg: `invalid kubernetes cluster UUID "not-a-uuid": invalid UUID length: 10`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			api := newFakeAPI(t, test.responses)
			conf := config.New()
			c := commands.BuildCommand(ScaleCommand(), nil, conf)
			c.Cobra().SetArgs(test.args)

			_, err := mockexecute.MockExecuteWithV9(c, &smock.Service{}, api.client, conf)

			if test.errorMsg != "" {
				assert.EqualError(t, err, test.errorMsg)
			} else {
				require.NoError(t, err)
			}

			requests := api.all()
			if test.body != "" {
				require.Len(t, requests, 1)
				assert.Equal(t, http.MethodPatch, requests[0].Method)
				assert.Equal(t, path, requests[0].Path)
				assert.JSONEq(t, test.body, string(requests[0].Body))
			} else {
				assert.Empty(t, requests)
			}
		})
	}
}
