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

func TestAccVPSResource(t *testing.T) {
	identifier := testAccIdentifier("tfvps", 20)
	resourceAddress := "mythicbeasts_vps." + identifier

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVPSResourceConfig(identifier),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("identifier"),
						knownvalue.StringExact(identifier),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("cpu_mode"),
						knownvalue.StringExact("performance"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("zone"),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"code": knownvalue.NotNull(),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ipv4"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("product"),
						knownvalue.StringExact("VPSX4"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         resourceAddress,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        identifier,
				ImportStateVerifyIdentifierAttribute: "identifier",
			},
			// Update and Read testing
			{
				Config: testAccVPSResourceConfig(identifier),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("identifier"),
						knownvalue.StringExact(identifier),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("name"),
						knownvalue.StringExact(identifier),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("disk_bus"),
						knownvalue.StringExact("virtio"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("zone"),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"code": knownvalue.NotNull(),
						}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ipv4"),
						knownvalue.SetSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("product"),
						knownvalue.StringExact("VPSX4"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVPSResourceConfig(identifier string) string {
	return fmt.Sprintf(`
resource "mythicbeasts_vps" %[1]q {
  identifier     = %[1]q
  name           = %[1]q
  disk_size      = 10240
  image          = "cloudinit-ubuntu-noble.raw.gz"
  ipv4_enabled   = false
  product        = "VPSX4"
  ssh_keys       = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
  create_in_zone = "uk"
}
`, identifier)
}
