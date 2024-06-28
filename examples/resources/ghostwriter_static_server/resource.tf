resource "ghostwriter_static_server" "test" {
  name               = "hostname"
  server_provider_id = 1
  ip_address         = "192.168.0.1"
  note               = "Test note"
}