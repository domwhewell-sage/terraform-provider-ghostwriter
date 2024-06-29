package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestStaticServerCheckoutResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
data "ghostwriter_activity_type" "test" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_static_server" "test" {
  name = "hostname"
  server_provider_id = 1
  ip_address = "192.168.0.1"
  note = "Test Note"
}

resource "ghostwriter_static_server_checkout" "test" {
  project_id       = data.ghostwriter_project.testproject.id
  server_id        = 1
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.test.id
  server_role_id = 1
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "project_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "server_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "end_date", "2025-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "activity_type_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "note", ""),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "server_role_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server_checkout.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server_checkout.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ghostwriter_static_server_checkout.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"force_delete", "last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
data "ghostwriter_activity_type" "test" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_static_server" "test" {
  name = "hostname"
  server_provider_id = 1
  ip_address = "192.168.0.1"
  note = "Test Note"
}

resource "ghostwriter_static_server_checkout" "test" {
  project_id       = data.ghostwriter_project.testproject.id
  server_id        = 1
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.test.id
  server_role_id = 1
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "project_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "server_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "end_date", "2025-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "activity_type_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "note", ""),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "server_role_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server_checkout.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server_checkout.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server_checkout.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
