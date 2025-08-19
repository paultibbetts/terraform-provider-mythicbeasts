terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

resource "mythicbeasts_pi" "four" {
  identifier   = "pangolin-on-a"
  disk_size = 20
  model        = 4
  os_image     = "rpi-bookworm-arm64"
  ssh_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
  wait_for_dns = true
  memory       = 4096
}

//resource "mythicbeasts_pi" "three" {
//  identifier   = "maybelast3attempt"
//  model        = 3
//  disk_size = 20
//  os_image     = "rpi-bookworm-arm64"
//  ssh_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
//  wait_for_dns = false
//  memory = 1024
//}

