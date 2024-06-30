data "ghostwriter_activity_type" "commandandcontrol" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "Test Project"
}

resource "ghostwriter_static_server" "test" {
  name               = "hostname"
  server_provider_id = 1
  ip_address         = "192.168.0.1"
  note               = "Test note"
}

resource "ghostwriter_static_server_checkout" "test" {
  project_id       = data.ghostwriter_project.testproject.id
  server_id        = resource.ghostwriter_static_server.test.id
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.test.id
  server_role_id   = 1
  force_delete     = true
}