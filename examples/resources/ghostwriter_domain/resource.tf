resource "ghostwriter_domain" "example" {
  name       = "example.com"
  registrar  = "GoDaddy"
  creation   = "2024-01-01"
  expiration = "2025-01-01"
  auto_renew = false
  note       = "Testing domain"
}