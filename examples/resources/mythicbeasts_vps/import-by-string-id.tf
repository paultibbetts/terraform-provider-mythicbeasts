import {
  to = mythicbeasts_vps.example
  id = "example"
}

resource "mythicbeasts_vps" "example" {
  identifier   = "example"
  name         = "example"
  disk_size    = 10240
  image        = "cloudinit-ubuntu-noble.raw.gz"
  product      = "VPSX4"
  ssh_keys     = "ssh-ed25519 ..."
  ipv4_enabled = false
}
