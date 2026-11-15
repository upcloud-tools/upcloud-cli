package nodegroup

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/UpCloudLtd/upcloud-cli/v3/internal/commands"
	upcloudv8 "github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	upcloudv9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
)

func v9Client(exec commands.Executor) (*upcloudv9.ClientWithResponses, error) {
	client := exec.V9()
	if client == nil {
		return nil, errors.New("UpCloud API v9 client is not available, run with --debug for details")
	}
	return client, nil
}

// legacyNodeGroupOutput preserves the node group JSON/YAML contract while the API call uses v9.
func legacyNodeGroupOutput(ng *upcloudv9.KubernetesNodeGroup) (*upcloudv8.KubernetesNodeGroup, error) {
	body, err := json.Marshal(ng)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal v9 node group response: %w", err)
	}
	legacy := &upcloudv8.KubernetesNodeGroup{}
	if err := json.Unmarshal(body, legacy); err != nil {
		return nil, fmt.Errorf("cannot convert v9 node group response to legacy output: %w", err)
	}
	return legacy, nil
}
