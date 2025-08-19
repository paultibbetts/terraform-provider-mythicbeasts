terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_images" "all" {}

output "all_images" {
  value = data.mythicbeasts_vps_images.all
}


