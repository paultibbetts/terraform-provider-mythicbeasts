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

func TestAccVPSProductsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccVPSProductsDataSourceConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_products.all",
						tfjsonpath.New("products"),
						knownvalue.SetPartial([]knownvalue.Check{}),
					),
					statecheck.ExpectKnownValue(
						"data.mythicbeasts_vps_products.all",
						tfjsonpath.New("products"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"code":        knownvalue.NotNull(),
								"description": knownvalue.NotNull(),
								"family":      knownvalue.NotNull(),
								"name":        knownvalue.NotNull(),
								"period":      knownvalue.NotNull(),
								"specs": knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"bandwidth": knownvalue.NotNull(),
									"cores":     knownvalue.NotNull(),
									"ram":       knownvalue.NotNull(),
								}),
							}),
						}),
					),
				},
			},
		},
	})
}

const testAccVPSProductsDataSourceConfig = `
data "mythicbeasts_vps_products" "all" {}
`
