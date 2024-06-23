package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCloudServerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "ghostwriter_activity_type" "commandandcontrol" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_cloud_server" "test" {
  server_provider_id = 1
  activity_type_id = data.ghostwriter_activity_type.commandandcontrol.id
  ip_address = "192.168.0.1"
  aux_address = ["192.168.0.1"]
  project_id = data.ghostwriter_project.testproject.id
  operator_id = 1
  server_role_id = 1
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "name", ""),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "server_provider_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "activity_type_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "ip_address", "192.168.0.1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "aux_address.#", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "aux_address.0", "192.168.0.1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "project_id", "9"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "note", ""),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "operator_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "server_role_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_cloud_server.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_cloud_server.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ghostwriter_cloud_server.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"force_delete", "last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "ghostwriter_activity_type" "commandandcontrol" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_cloud_server" "test" {
  server_provider_id = 1
  activity_type_id = data.ghostwriter_activity_type.commandandcontrol.id
  ip_address = "192.168.0.2"
  project_id = data.ghostwriter_project.testproject.id
  operator_id = 1
  server_role_id = 1
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "name", ""),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "server_provider_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "activity_type_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "ip_address", "192.168.0.2"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "aux_address.#", "0"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "project_id", "9"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "note", ""),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "operator_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "server_role_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_cloud_server.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_cloud_server.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_cloud_server.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
