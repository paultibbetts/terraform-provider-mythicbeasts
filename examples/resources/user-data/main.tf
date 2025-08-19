terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

resource "mythicbeasts_user_data" "example" {
	name = "example-apache"
	data = "#cloud-config\n\npackages:\n  - apache2\n\n"
}
