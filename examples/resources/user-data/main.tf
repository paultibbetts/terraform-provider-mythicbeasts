terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

provider "mythicbeasts" {
  keyid  = "wmmncd9gpha8vk8p"
  secret = "DVFekI5rfWMeF5BdjCylY60cBAZW87"
}

resource "mythicbeasts_user_data" "example" {
	name = "example-apache"
	data = "#cloud-config\n\npackages:\n  - apache2\n\n"
}
