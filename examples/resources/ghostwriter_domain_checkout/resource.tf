data "ghostwriter_activity_type" "commandandcontrol" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "Test Project"
}

resource "ghostwriter_domain" "example" {
  name       = "example.com"
  registrar  = "GoDaddy"
  creation   = "2024-01-01"
  expiration = "2025-01-01"
  auto_renew = false
  note       = "Testing domain"
}

resource "ghostwriter_domain_checkout" "example" {
  project_id       = data.ghostwriter_project.testproject.id
  domain_id        = resource.ghostwriter_domain.example.id
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.commandandcontrol.id
  note             = "Example Note"
}