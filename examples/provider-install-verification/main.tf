terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.1.0"
    }
  }
}

provider "mythicbeasts" {}

data "mythicbeasts_pis" "example" {}

output "example" {
  value = data.mythicbeasts_pis.example
}
