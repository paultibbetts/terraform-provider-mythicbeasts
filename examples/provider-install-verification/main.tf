terraform {
  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.1.0"
    }
  }
}

provider "mythicbeasts" {}

data "mythicbeasts_pis" "example" {}

