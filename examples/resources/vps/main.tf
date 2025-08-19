terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

resource "mythicbeasts_vps" "test" {
  identifier     = "paulsvpsinuk"
  name           = "paulsvpsinuk"
  disk_size      = 10240
  image          = "cloudinit-ubuntu-noble.raw.gz"
  ipv4_enabled   = false
  product        = "VPSX4"
  ssh_keys       = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
  create_in_zone = "uk"
}

