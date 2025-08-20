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

var userData = "#cloud-config\n\npackages:\n  - apache2\n\n"
var size = int64(len(userData))

func TestAccUserDataResource(t *testing.T) {
	t.Setenv("TF_LOG", "INFO")
	t.Setenv("TF_LOG_PATH", "-")
	t.Setenv("TF_ACC_TERRAFORM_LOG_PATH", "-")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserDataResourceConfig("web-server"),
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
						knownvalue.StringExact(userData),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_user_data.web-server",
						tfjsonpath.New("size"),
						knownvalue.Int64Exact(size),
					),
				},
			},
			// ImportState testing
			//{
			//	ResourceName:      "mythicbeasts_user_data.web-server",
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
			// Update and Read testing
			//{},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserDataResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "mythicbeasts_user_data" %[1]q {
  name           = %[1]q
  data      = "#cloud-config\n\npackages:\n  - apache2\n\n"
}
`, name)
}
