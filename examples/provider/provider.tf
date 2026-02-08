terraform {
  required_version = ">= 1.11.0"

  required_providers {
    mythicbeasts = {
      source = "paultibbetts/mythicbeasts"
    }
  }
}

provider "mythicbeasts" {
  # Leave unset to use environment variables:
  # MYTHICBEASTS_KEYID and MYTHICBEASTS_SECRET
  #
  # keyid  = "your-keyid"
  # secret = "your-secret"
}
