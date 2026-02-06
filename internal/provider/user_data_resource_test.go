// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var testAccUserData = "#cloud-config\n\npackages:\n  - apache2\n\n"
var testAccUserDataSize = int64(len(testAccUserData))
var testAccUserDataUpdated = "#cloud-config\n\npackages:\n  - nginx\n\n"
var testAccUserDataUpdatedSize = int64(len(testAccUserDataUpdated))

func TestAccUserDataResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserDataResourceConfig("web-server", testAccUserData),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("name"),
						knownvalue.StringExact("web-server"),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("data"),
						knownvalue.StringExact(testAccUserData),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("size"),
						knownvalue.Int64Exact(testAccUserDataSize),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "mythicbeasts_user_data.web-server",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccUserDataResourceConfig("web-server", testAccUserDataUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("name"),
						knownvalue.StringExact("web-server"),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("data"),
						knownvalue.StringExact(testAccUserDataUpdated),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("size"),
						knownvalue.Int64Exact(testAccUserDataUpdatedSize),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserDataResourceConfig(name, data string) string {
	return fmt.Sprintf(`
resource "mythicbeasts_user_data" %[1]q {
  name           = %[1]q
  data           = %[2]q
}
`, name, data)
}
