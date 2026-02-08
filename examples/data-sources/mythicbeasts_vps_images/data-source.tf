terraform {
  required_version = ">= 1.11.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_images" "all" {}

output "all_images" {
  value = data.mythicbeasts_vps_images.all
}


