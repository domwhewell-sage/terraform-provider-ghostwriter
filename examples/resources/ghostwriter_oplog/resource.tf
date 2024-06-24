data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_oplog" "test" {
  name               = "Test Oplog"
  project_id         = data.ghostwriter_project.testproject.id
  mute_notifications = true
  force_delete       = true
}