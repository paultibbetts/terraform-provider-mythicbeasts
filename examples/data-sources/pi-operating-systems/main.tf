terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.2.0"
    }
  }
}

data "mythicbeasts_pi_operating_systems" "three" {
  model = 3
}

output "os_for_model_three" {
  value = data.mythicbeasts_pi_operating_systems.three
}

