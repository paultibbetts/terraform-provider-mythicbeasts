// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccVPSZonesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccVPSZonesDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_zones.all",
						tfjsonpath.New("zones"),
						knownvalue.SetPartial([]knownvalue.Check{}),
					),
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_zones.all",
						tfjsonpath.New("zones"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"description": knownvalue.NotNull(),
								"name":        knownvalue.NotNull(),
								"parents": knownvalue.ListPartial(map[int]knownvalue.Check{
									0: knownvalue.NotNull(),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

const testAccVPSZonesDataSourceConfig = `
data "mythicbeasts_vps_zones" "all" {}
`
