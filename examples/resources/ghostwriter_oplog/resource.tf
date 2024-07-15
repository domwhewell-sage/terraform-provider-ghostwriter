data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_oplog" "test" {
  name         = "Test Oplog"
  project_id   = data.ghostwriter_project.testproject.id
  force_delete = true
}