import {
  to = mythicbeasts_user_data.example
  id = "123"
}

resource "mythicbeasts_user_data" "example" {
  name = "example-apache"
  data = "#cloud-config\n\npackages:\n  - apache2\n\n"
}
