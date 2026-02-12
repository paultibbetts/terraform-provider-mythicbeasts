// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const (
	proxyEndpointHostname     = "example"
	proxyEndpointSite         = "all"
	proxyEndpointDomainEnvVar = "MB_TEST_PROXY_DOMAIN"
)

func TestAccProxyEndpointResource(t *testing.T) {
	proxyEndpointPiIdentifier := testAccIdentifier("tfpx", 20)
	domain := testAccProxyEndpointDomain(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyEndpointResourceConfig(proxyEndpointPiIdentifier, domain, proxyEndpointHostname, proxyEndpointSite, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mythicbeasts_proxy_endpoint.test",
						tfjsonpath.New("domain"),
						knownvalue.StringExact(domain),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_proxy_endpoint.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(proxyEndpointHostname),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_proxy_endpoint.test",
						tfjsonpath.New("site"),
						knownvalue.StringExact(proxyEndpointSite),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_proxy_endpoint.test",
						tfjsonpath.New("proxy_protocol"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"mythicbeasts_proxy_endpoint.test",
						tfjsonpath.New("address"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccProxyEndpointResourceConfig(proxyEndpointPiIdentifier, domain, proxyEndpointHostname, proxyEndpointSite, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mythicbeasts_proxy_endpoint.test",
						tfjsonpath.New("proxy_protocol"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccProxyEndpointResourceConfig(proxyEndpointPiIdentifier, domain, proxyEndpointHostname, proxyEndpointSite, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"mythicbeasts_proxy_endpoint.test",
						tfjsonpath.New("proxy_protocol"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:      "mythicbeasts_proxy_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rs, ok := state.RootModule().Resources["mythicbeasts_proxy_endpoint.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}

					domain := rs.Primary.Attributes["domain"]
					hostname := rs.Primary.Attributes["hostname"]
					address := rs.Primary.Attributes["address"]
					site := rs.Primary.Attributes["site"]
					if domain == "" || hostname == "" || address == "" || site == "" {
						return "", fmt.Errorf("missing import ID components in state")
					}

					return fmt.Sprintf("%s/%s/%s/%s", domain, hostname, address, site), nil
				},
			},
		},
	})
}

func testAccProxyEndpointDomain(t *testing.T) string {
	t.Helper()

	domain := os.Getenv(proxyEndpointDomainEnvVar)
	if domain == "" {
		t.Skipf("set %s to run proxy endpoint acceptance tests", proxyEndpointDomainEnvVar)
	}

	return domain
}

func testAccProxyEndpointResourceConfig(identifier, domain, hostname, site string, proxyProtocol bool) string {
	return fmt.Sprintf(`
resource "mythicbeasts_pi" "proxy" {
  identifier   = %[1]q
  disk_size    = 10
  model        = 4
  os_image     = "rpi-bookworm-arm64"
  ssh_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
  wait_for_dns = true
  memory       = 4096
}

resource "mythicbeasts_proxy_endpoint" "test" {
  domain         = %[2]q
  hostname       = %[3]q
  address        = mythicbeasts_pi.proxy.ip
  site           = %[4]q
  proxy_protocol = %[5]t
}
`, identifier, domain, hostname, site, proxyProtocol)
}
