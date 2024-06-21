terraform {
  required_providers {
    ghostwriter = {
      source = "example.com/domwhewell/ghostwriter"
    }
  }
}

provider "ghostwriter" {
  endpoint = "http://localhost:8080/v1/graphql"
  api_key  = "ey..."
}

resource "ghostwriter_domain" "example" {
  name               = "example.com"
  registrar          = "amazon"
  creation           = "2024-01-01"
  expiration         = "2025-01-01"
  auto_renew         = false
  note               = "This is an example domain"
  burned_explanation = ""
  vt_permalink       = ""
}

output "example_domain" {
  value = ghostwriter_domain.example
}