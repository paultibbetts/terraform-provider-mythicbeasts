# Terraform Provider for Mythic Beasts

Terraform provider for the Mythic Beasts [Pi](https://www.mythic-beasts.com/support/api/raspberry-pi), [Proxy](https://www.mythic-beasts.com/support/api/proxy) and [VPS](https://www.mythic-beasts.com/support/api/vps) APIs.

## Table of Contents

- [Requirements](#requirements)
- [Authentication](#authentication)
- [Proxy endpoint domain management](#proxy-endpoint-domain-management)
- [Installation](#installation)
- [Example Usage](#example-usage)
- [Resources](#resources)
- [Data sources](#data-sources)
- [Development Status](#development-status)
- [Versioning](#versioning)
- [Building The Provider](#building-the-provider)
- [Adding Dependencies](#adding-dependencies)
- [Developing the Provider](#developing-the-provider)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.11.0
- [Go](https://golang.org/doc/install) >= 1.24

## Authentication

The Mythic Beasts APIs require an [API key](https://www.mythic-beasts.com/customer/api-users) for authentication.

When creating the key you must add permissions to work with the APIs you wish to use, such as:

- "Virtual Server Provisioning" for VPS
- "Raspberry Pi provisioning" for Pis
- "IPv4 to IPv6 Proxy API" for Proxy Endpoints

## Proxy endpoint domain management

The domain used for proxy endpoints must be registered with the Mythic Beasts control panel.

This can be done by registering the domain using their [domain management](https://www.mythic-beasts.com/customer/domains) or by [adding it as a 3rd party domain](https://www.mythic-beasts.com/customer/3rdpartydomain).

## Installation

```hcl
terraform {
  required_version = ">= 1.11.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

provider "mythicbeasts" {
  keyid  = "your-keyid"
  secret = "your-secret"
}
```

Alternatively, credentials can be supplied via environment variables: 

- `MYTHICBEASTS_KEYID`
- `MYTHICBEASTS_SECRET`

## Example Usage

```hcl
terraform {
  required_version = ">= 1.11.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

resource "mythicbeasts_pi" "server" {
  identifier = "example"
  disk_size  = 10
  model      = 4
  memory     = 4096
}

resource "mythicbeasts_proxy_endpoint" "apex" {
  domain         = "example.com"
  hostname       = "@"
  address        = mythicbeasts_pi.server.ip
  site           = "all"
  proxy_protocol = true
}

resource "mythicbeasts_proxy_endpoint" "www" {
  domain         = "example.com"
  hostname       = "www"
  address        = mythicbeasts_pi.server.ip
  site           = "all"
  proxy_protocol = true
}
```

For more see [`examples/`](./examples).

## Resources

- `mythicbeasts_pi` - Raspberry Pi (IPv6 only)
- `mythicbeasts_proxy_endpoint` - IPv4-IPv6 proxy endpoint
- `mythicbeasts_vps` - Virtual Private Server

## Data sources

- `mythicbeasts_pi_models` - Raspberry Pi server models and specs
- `mythicbeasts_pi_operating_systems` - Raspberry Pi operating system images

- `mythicbeasts_vps_disk_sizes` - VPS disk sizes
- `mythicbeasts_vps_hosts` - VPS private cloud host servers
- `mythicbeasts_vps_images` - VPS operating system images
- `mythicbeasts_vps_pricing` - VPS pricing information
- `mythicbeasts_vps_products` - VPS products
- `mythicbeasts_vps_zones` - VPS zones (datacentres)

## Development status

This is a community-maintained provider and is not affiliated with or endorsed by Mythic Beasts. 

It is provided on a best-effort basis. There is no formal support offering.

Issues and pull requests are welcome.

## Versioning

This project is pre-1.0 and minor releases may include breaking changes.

[Semantic versioning](https://semver.org/) is used for tags and a v1.0.0 will signal a stable API.

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and cost money to run.

```shell
make testacc
```

The `mythicbeasts_proxy_endpoint` resource requires the domain to be registered with the Mythic Beasts control panel.

This domain could be registered with Mythic Beasts using their [domain management](https://www.mythic-beasts.com/customer/domains) or [added as a 3rd party domain](https://www.mythic-beasts.com/customer/3rdpartydomain).

To run the acceptance tests for the `mythicbeasts_proxy_endpoint` resource you must also pass in the domain to use as `MB_TEST_PROXY_DOMAIN`.

```shell
MB_TEST_PROXY_DOMAIN=example.com make testacc
```
