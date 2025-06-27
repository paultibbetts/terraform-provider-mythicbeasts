terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

provider "mythicbeasts" {}

data "mythicbeasts_pis" "example" {}

