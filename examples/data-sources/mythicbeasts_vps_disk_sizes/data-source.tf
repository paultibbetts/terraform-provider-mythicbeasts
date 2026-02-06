terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_disk_sizes" "all" {}

output "all_disk_sizes" {
  value = data.mythicbeasts_vps_disk_sizes.all
}


