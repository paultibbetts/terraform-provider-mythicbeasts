terraform {
  required_providers {
    mythicbeasts = {
      source = "paultibbetts.uk/terraform-providers/mythicbeasts"
    }
  }
}

data "mythicbeasts_pi_operating_systems" "three" {
  model = 3
}

output "os_for_model_three" {
  value = data.mythicbeasts_pi_operating_systems.three
}

