resource "ghostwriter_domain_server_connection" "test" {
  domain_checkout_id = 1
  project_id         = 1
  cloud_server_id    = 1
  subdomain          = "*"
  endpoint           = "/login"
  force_delete       = true
}