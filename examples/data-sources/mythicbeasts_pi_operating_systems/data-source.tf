terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

data "mythicbeasts_pi_operating_systems" "four" {
  model = 4
}

output "os_for_model_four" {
  value = data.mythicbeasts_pi_operating_systems.four
}

