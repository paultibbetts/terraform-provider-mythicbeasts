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

func TestAccVPSDiskSizesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccVPSDiskSizesDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_disk_sizes.all",
						tfjsonpath.New("hdd"),
						knownvalue.ListPartial(map[int]knownvalue.Check{}),
					),
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_disk_sizes.all",
						tfjsonpath.New("hdd"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"size": knownvalue.NotNull(),
							}),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_disk_sizes.all",
						tfjsonpath.New("ssd"),
						knownvalue.ListPartial(map[int]knownvalue.Check{}),
					),
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_disk_sizes.all",
						tfjsonpath.New("ssd"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"size": knownvalue.NotNull(),
							}),
						}),
					),
				},
			},
		},
	})
}

const testAccVPSDiskSizesDataSourceConfig = `
data "mythicbeasts_vps_disk_sizes" "all" {}
`
