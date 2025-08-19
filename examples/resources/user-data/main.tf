terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.1.0"
    }
  }
}

resource "mythicbeasts_user_data" "example" {
	name = "example-apache"
	data = "#cloud-config\n\npackages:\n  - apache2\n\n"
}
