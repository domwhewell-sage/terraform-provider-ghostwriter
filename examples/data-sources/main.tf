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

data "ghostwriter_activity_type" "commandandcontrol" {
  name = "Command"
}

output "activity_type" {
  value = data.ghostwriter_activity_type.commandandcontrol
}