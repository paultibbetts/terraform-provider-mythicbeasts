// Copyright IBM Corp. 2021, 2026
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
	name := "web-server-" + testAccRunSuffix()
	resourceAddress := "mythicbeasts_user_data." + name

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserDataResourceConfig(name, testAccUserData),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("data"),
						knownvalue.StringExact(testAccUserData),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("size"),
						knownvalue.Int64Exact(testAccUserDataSize),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccUserDataResourceConfig(name, testAccUserDataUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("data"),
						knownvalue.StringExact(testAccUserDataUpdated),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
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
