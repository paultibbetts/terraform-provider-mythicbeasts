// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVPSPricingDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccVPSPricingDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_pricing.all",
						tfjsonpath.New("ipv4"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_pricing.all",
						tfjsonpath.New("disk"),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"hdd": knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"price":  knownvalue.NotNull(),
								"extent": knownvalue.NotNull(),
							}),
							"ssd": knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"price":  knownvalue.NotNull(),
								"extent": knownvalue.NotNull(),
							}),
						}),
					),
				},
			},
		},
	})
}

const testAccVPSPricingDataSourceConfig = `
data "mythicbeasts_vps_pricing" "all" {}
`
