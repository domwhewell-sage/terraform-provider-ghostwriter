package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOplogResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_oplog" "test" {
  name = "Test Oplog"
  project_id = data.ghostwriter_project.testproject.id
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "name", "Test Oplog"),
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "project_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "mute_notifications", "false"),
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_oplog.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_oplog.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ghostwriter_oplog.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"force_delete", "last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_oplog" "test" {
  name = "Test Updated Oplog"
  project_id = data.ghostwriter_project.testproject.id
  mute_notifications = true
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "name", "Test Updated Oplog"),
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "project_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "mute_notifications", "true"),
					resource.TestCheckResourceAttr("ghostwriter_oplog.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_oplog.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_oplog.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
