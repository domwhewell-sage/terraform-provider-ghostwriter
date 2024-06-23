resource "ghostwriter_cloud_server" "test" {
  name               = "hostname"
  server_provider_id = 1
  activity_type_id   = 1
  ip_address         = "192.168.0.1"
  aux_address        = ["192.168.0.2"]
  project_id         = 1
  note               = "test note"
  operator_id        = 1
  server_role_id     = 1
  force_delete       = true
}