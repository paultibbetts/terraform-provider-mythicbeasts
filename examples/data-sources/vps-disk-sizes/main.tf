terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

data "mythicbeasts_vps_disk_sizes" "all" {}

output "all_disk_sizes" {
	value = data.mythicbeasts_vps_disk_sizes.all
}


