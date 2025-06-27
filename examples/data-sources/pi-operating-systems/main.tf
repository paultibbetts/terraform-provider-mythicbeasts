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

data "mythicbeasts_pi_operating_systems" "three" {
  model = 3
}

output "os_for_model_three" {
  value = data.mythicbeasts_pi_operating_systems.three
}

