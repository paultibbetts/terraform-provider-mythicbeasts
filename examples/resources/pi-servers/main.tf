terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

provider "mythicbeasts" {
  keyid  = "wmmncd9gpha8vk8p"
  secret = "DVFekI5rfWMeF5BdjCylY60cBAZW87"
}

resource "mythicbeasts_pi_server" "four" {
  identifier   = "pi4createdbytf"
  model        = 4
  os_image     = "rpi-bookworm-arm64"
  ssh_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
  wait_for_dns = true
  memory       = 4096
}

resource "mythicbeasts_pi_server" "three" {
  identifier   = "pi3createdbytf"
  model        = 3
  os_image     = "rpi-bookworm-arm64"
  ssh_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPfx70ArvHPF+9U3GgKgNEAWkXSyZMun83sn9582Pl4e code@paultibbetts.uk"
  wait_for_dns = true
  memory       = 1024
}

