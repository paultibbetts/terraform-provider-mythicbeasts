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

const piSSHKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e"
const piSSHKeyUpdated = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB5Y6QFrfcJw0f0D8uG2D8btJQH8k8K6Pp9g8b0Xv2zL"

func TestAccPiResource(t *testing.T) {
	piIdentifier := testAccIdentifier("tfpi", 20)
	resourceAddress := "mythicbeasts_pi." + piIdentifier

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPiResourceConfig(piIdentifier, piSSHKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("identifier"),
						knownvalue.StringExact(piIdentifier),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("disk_size"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ip"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("model"),
						knownvalue.Int64Exact(4),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("memory"),
						knownvalue.Int64Exact(4096),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         resourceAddress,
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
				Config: testAccPiResourceConfig(piIdentifier, piSSHKeyUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("identifier"),
						knownvalue.StringExact(piIdentifier),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("disk_size"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("ip"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("model"),
						knownvalue.Int64Exact(4),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("memory"),
						knownvalue.Int64Exact(4096),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPiResourceConfig(identifier string, sshKey string) string {
	return fmt.Sprintf(`
resource "mythicbeasts_pi" %[1]q {
  identifier   = %[1]q
  disk_size    = 10
  model        = 4
  os_image     = "rpi-bookworm-arm64"
  ssh_key      = %[2]q
  wait_for_dns = true
  memory       = 4096
}
`, identifier, sshKey)
}
