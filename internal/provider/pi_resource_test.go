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

const piIdentifier = "tfprovidertest6"

func TestAccPiResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPiResourceConfig(piIdentifier),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("identifier"),
						knownvalue.StringExact(piIdentifier),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("disk_size"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("ip"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("model"),
						knownvalue.Int64Exact(4),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("memory"),
						knownvalue.Int64Exact(4096),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         "mythicbeasts_pi." + piIdentifier,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        piIdentifier,
				ImportStateVerifyIdentifierAttribute: "identifier",
				ImportStateVerifyIgnore: []string{
					"os_image",
					"wait_for_dns",
				},
			},
			// Update and Read testing
			{
				Config: testAccPiResourceConfig(piIdentifier),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("identifier"),
						knownvalue.StringExact(piIdentifier),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("disk_size"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("ip"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("model"),
						knownvalue.Int64Exact(4),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_pi."+piIdentifier,
						tfjsonpath.New("memory"),
						knownvalue.Int64Exact(4096),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPiResourceConfig(identifier string) string {
	return fmt.Sprintf(`
resource "mythicbeasts_pi" %[1]q {
  identifier   = %[1]q
  disk_size    = 10
  model        = 4
  os_image     = "rpi-bookworm-arm64"
  ssh_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
  wait_for_dns = true
  memory       = 4096
}
`, identifier)
}
