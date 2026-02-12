// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// Helper function to find a specific resource in state.
func findResource(st *tfjson.State, addr string) *tfjson.StateResource {
	if st == nil || st.Values == nil || st.Values.RootModule == nil {
		return nil
	}
	for _, r := range st.Values.RootModule.Resources {
		if r.Address == addr {
			return r
		}
	}
	return nil
}

type expectAllModelsEqual struct {
	resourceAddress string
	want            int64
}

func (e expectAllModelsEqual) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	res := findResource(req.State, e.resourceAddress)
	if res == nil {
		resp.Error = fmt.Errorf("%s - resource not found", e.resourceAddress)
		return
	}

	values, err := tfjsonpath.Traverse(res.AttributeValues, tfjsonpath.New("models"))
	if err != nil {
		resp.Error = err
		return
	}

	items, ok := values.([]any)
	if !ok {
		resp.Error = fmt.Errorf("models is not a list")
		return
	}
	if len(items) == 0 {
		resp.Error = fmt.Errorf("expected at least one model")
		return
	}

	for i, it := range items {
		obj, ok := it.(map[string]any)
		if !ok {
			resp.Error = fmt.Errorf("models[%d] is not an object", i)
			return
		}

		raw, ok := obj["model"]
		if !ok {
			resp.Error = fmt.Errorf("models[%d] missing 'model' field", i)
			return
		}

		var got int64
		switch n := raw.(type) {
		case int64:
			got = n
		case float64:
			got = int64(n)
		case json.Number:
			got, err = n.Int64()
			if err != nil {
				resp.Error = fmt.Errorf("error parsing json.Number as int %#v", raw)
				return
			}
		case string:
			got, err = strconv.ParseInt(n, 10, 64)
			if err != nil {
				resp.Error = fmt.Errorf("error parsing string as int %#v", raw)
				return
			}
		default:
			resp.Error = fmt.Errorf("models[%d].model has unexpected type %T (value=%#v)", i, raw, raw)
			return
		}

		if got != e.want {
			resp.Error = fmt.Errorf("models[%d].model=%d, want %d", i, got, e.want)
			return
		}
	}
}

func ExpectAllModelsEqual(addr string, want int64) statecheck.StateCheck {
	return expectAllModelsEqual{resourceAddress: addr, want: want}
}

func TestAccPiModelsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccPiModels3DataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectAllModelsEqual("data.mythicbeasts_pi_models.test", 3),
				},
			},
			{
				Config: testAccPiModels4With4GBRamDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_pi_models.four_with_four_gb_ram",
						tfjsonpath.New("models"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"model":  knownvalue.Int64Exact(4),
								"memory": knownvalue.Int64Exact(4096),
							}),
						}),
					),
				},
			},
		},
	})
}

const testAccPiModels3DataSourceConfig = `
data "mythicbeasts_pi_models" "test" {
  model = 3
}
`
const testAccPiModels4With4GBRamDataSourceConfig = `
data "mythicbeasts_pi_models" "four_with_four_gb_ram" {
  model  = 4
  memory = 4096
}
`
