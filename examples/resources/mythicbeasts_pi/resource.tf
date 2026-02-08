terraform {
  required_version = ">= 1.11.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

resource "mythicbeasts_pi" "example" {
  identifier   = "example"
  disk_size    = 10
  model        = 4
  os_image     = "rpi-bookworm-arm64"
  ssh_key      = "ssh-ed25519 ..."
  wait_for_dns = true
  memory       = 4096
}

