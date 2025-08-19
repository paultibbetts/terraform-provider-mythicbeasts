terraform {
  required_version = ">= 1.3.0"

  required_providers {
    mythicbeasts = {
      source  = "paultibbetts.uk/terraform-providers/mythicbeasts"
      version = "~> 0.1.0"
    }
  }
}

data "mythicbeasts_pi_models" "all" {}

data "mythicbeasts_pi_models" "three" {
  model = 3
}

data "mythicbeasts_pi_models" "fastest" {
  cpu_speed = 2000
}

data "mythicbeasts_pi_models" "four_with_four_gb_ram" {
  model  = 4
  memory = 4096
}

output "pi_models" {
  value = data.mythicbeasts_pi_models.all.models
}

output "pi_3_models" {
  value = data.mythicbeasts_pi_models.three.models
}

output "pi_fastest_model" {
  value = data.mythicbeasts_pi_models.fastest.models[0]
}

output "pi_four_with_four_gb_ram_model" {
  value = data.mythicbeasts_pi_models.four_with_four_gb_ram.models[0]
}


