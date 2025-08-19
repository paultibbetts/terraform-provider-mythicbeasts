terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.1.0"
    }
  }
}

data "mythicbeasts_vps_products" "all" {}

output "vps_products" {
  value = data.mythicbeasts_vps_products.all
}

