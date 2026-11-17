package server

import (
	"encoding/json"
	"testing"

	"github.com/UpCloudLtd/upcloud-cli/v3/internal/commands"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/config"
	smock "github.com/UpCloudLtd/upcloud-cli/v3/internal/mock"
	"github.com/UpCloudLtd/upcloud-cli/v3/internal/mockexecute"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationHeader(t *testing.T) {
	testCases := []struct {
		input          string
		expectedHeader string
	}{
		{
			input:          "hour",
			expectedHeader: "Price (per hour)",
		},
		{
			input:          "month",
			expectedHeader: "Price (per month)",
		},
		{
			input:          "1h",
			expectedHeader: "Price (per hour)",
		},
		{
			input:          "4h30m",
			expectedHeader: "Price (per 4h30m)",
		},
		{
			input:          "24h",
			expectedHeader: "Price (per day)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			header := formatPricingHeader(tc.input)
			assert.Equal(t, tc.expectedHeader, header)
		})
	}
}

func TestGetPlanCost(t *testing.T) {
	prices := map[string]upcloud.Price{
		"server_plan_1xCPU-1GB": {
			Amount: 1,
			Price:  1.0416,
		},
	}

	testcases := []struct {
		plan     upcloud.Plan
		duration string
		expected float64
	}{
		{
			plan:     upcloud.Plan{Name: "1xCPU-1GB"},
			duration: "1h",
			expected: 0.010416,
		},
		{
			plan:     upcloud.Plan{Name: "1xCPU-1GB"},
			duration: "2h30m",
			expected: 0.031248,
		},
		{
			plan:     upcloud.Plan{Name: "1xCPU-1GB"},
			duration: "hour",
			expected: 0.010416,
		},
		{
			plan:     upcloud.Plan{Name: "1xCPU-1GB"},
			duration: "month",
			expected: 7,
		},
	}

	for _, tc := range testcases {
		name := tc.plan.Name + "-" + tc.duration
		t.Run(name, func(t *testing.T) {
			duration, err := getDuration(tc.duration)
			assert.NoError(t, err)

			cost := getPlanCost(tc.plan, prices, duration)
			assert.InDelta(t, tc.expected, cost, 0.001)
		})
	}
}

func TestPlanListMachineOutput(t *testing.T) {
	plans := &upcloud.Plans{
		Plans: []upcloud.Plan{
			{Name: "1xCPU-1GB", CoreNumber: 1, MemoryAmount: 1024, StorageSize: 25, StorageTier: "maxiops", PublicTrafficOut: 1024},
			{Name: "PREMIUM-2xCPU-2GB", CoreNumber: 2, MemoryAmount: 2048, StorageSize: 50, StorageTier: "maxiops", PublicTrafficOut: 2048},
			{Name: "DEV-1xCPU-1GB", CoreNumber: 1, MemoryAmount: 1024, StorageSize: 25, StorageTier: "maxiops", PublicTrafficOut: 1024},
		},
	}
	prices := upcloud.PricesByZone{
		"de-fra1": map[string]upcloud.Price{
			"server_plan_1xCPU-1GB":         {Amount: 1, Price: 100},
			"server_plan_PREMIUM-2xCPU-2GB": {Amount: 1, Price: 200},
		},
	}

	for _, test := range []struct {
		name           string
		args           []string
		expectedPrices map[string]float64
	}{
		{
			name:           "no prices requested",
			args:           []string{},
			expectedPrices: map[string]float64{},
		},
		{
			name:           "hourly prices",
			args:           []string{"--prices", "de-fra1", "--prices-duration", "hour"},
			expectedPrices: map[string]float64{"1xCPU-1GB": 1, "PREMIUM-2xCPU-2GB": 2},
		},
		{
			name:           "monthly prices",
			args:           []string{"--prices", "de-fra1"},
			expectedPrices: map[string]float64{"1xCPU-1GB": 672, "PREMIUM-2xCPU-2GB": 1344},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			mService := new(smock.Service)
			mService.On("GetPlans").Return(plans, nil)
			mService.On("GetPricesByZone").Return(&prices, nil)

			conf := config.New()
			conf.Viper().Set(config.KeyOutput, config.ValueOutputJSON)

			command := commands.BuildCommand(PlanListCommand(), nil, conf)
			command.Cobra().SetArgs(test.args)
			out, err := mockexecute.MockExecute(command, mService, conf)
			assert.NoError(t, err)

			var parsed []map[string]any
			require.NoError(t, json.Unmarshal([]byte(out), &parsed))
			require.Len(t, parsed, len(plans.Plans))
			for _, p := range parsed {
				name, _ := p["name"].(string)
				assert.Contains(t, p, "core_number", "plan %s should keep API fields", name)
				expected, wantPrice := test.expectedPrices[name]
				price, hasPrice := p["price"]
				if wantPrice {
					assert.True(t, hasPrice, "plan %s should include price", name)
					assert.Equal(t, expected, price)
				} else {
					assert.False(t, hasPrice, "plan %s should not include price", name)
				}
			}
		})
	}
}
