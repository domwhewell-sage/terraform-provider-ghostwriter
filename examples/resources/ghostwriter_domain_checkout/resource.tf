resource "ghostwriter_domain_checkout" "example" {
  project_id       = 1
  domain_id        = 1
  start_date       = "2024-01-01"
  end_date         = "2025-01-01"
  activity_type_id = 1
  note             = "Example Note"
}