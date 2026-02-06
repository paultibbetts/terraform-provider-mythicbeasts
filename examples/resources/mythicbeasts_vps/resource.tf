terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

resource "mythicbeasts_vps" "example" {
  identifier     = "example"
  name           = "example"
  disk_size      = 10240
  image          = "cloudinit-ubuntu-noble.raw.gz"
  ipv4_enabled   = false
  product        = "VPSX4"
  ssh_keys       = "ssh-ed25519 ..."
  create_in_zone = "uk"
}

