import {
  to = mythicbeasts_proxy_endpoint.example
  id = "example.com/example/2001:db8::1/all"
}

resource "mythicbeasts_proxy_endpoint" "example" {
  domain   = "example.com"
  hostname = "example"
  address  = "2001:db8::1"
  site     = "all"
}
