package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDomainServerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "ghostwriter_activity_type" "test" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_domain" "test" {
  name = "test.com"
  creation = "2024-01-01"
  expiration = "2025-01-01"
  force_delete = true
}

resource "ghostwriter_domain_checkout" "test" {
  project_id       = data.ghostwriter_project.testproject.id
  domain_id        = resource.ghostwriter_domain.test.id
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.test.id
  force_delete = true
}

resource "ghostwriter_static_server_checkout" "test" {
  project_id       = 1
  server_id        = 1
  start_date       = "2024-01-01"
  end_date         = "2025-01-01"
  activity_type_id = 1
  server_role_id = 1
  note = "Test Note"
  force_delete = true
}

resource "ghostwriter_domain_server" "test" {
  domain_checkout_id       = resource.ghostwriter_domain_checkout.test.id
  project_id        = data.ghostwriter_project.testproject.id
  static_server_checkout_id       = resource.ghostwriter_static_server_checkout.test.id
  subdomain         = "login"
  endpoint = "/test"
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "domain_checkout_id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "project_id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "static_server_checkout_id"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "subdomain", "login"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "endpoint", "/test"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "last_updated"),
				),
			},
			// Create and Read testing
			{
				Config: providerConfig + `
data "ghostwriter_activity_type" "test" {
  name = "Command and Control"
}

data "ghostwriter_project" "testproject" {
  code_name = "TestProject"
}

resource "ghostwriter_domain" "test" {
  name = "test.com"
  creation = "2024-01-01"
  expiration = "2025-01-01"
  force_delete = true
}

resource "ghostwriter_domain_checkout" "test" {
  project_id       = data.ghostwriter_project.testproject.id
  domain_id        = resource.ghostwriter_domain.test.id
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.test.id
  force_delete = true
}

resource "ghostwriter_cloud_server" "test" {
  server_provider_id = 1
  activity_type_id = data.ghostwriter_activity_type.test.id
  ip_address = "192.168.0.1"
  aux_address = ["192.168.0.1"]
  project_id = data.ghostwriter_project.testproject.id
  server_role_id = 1
  force_delete = true
}

resource "ghostwriter_domain_server" "test" {
  domain_checkout_id       = resource.ghostwriter_domain_checkout.test.id
  project_id        = data.ghostwriter_project.testproject.id
  cloud_server_id       = resource.ghostwriter_cloud_server.test.id
  subdomain         = "login"
  endpoint = "/test"
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "domain_checkout_id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "project_id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "cloud_server_id"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "subdomain", "login"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "endpoint", "/test"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ghostwriter_domain_server.test",
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

resource "ghostwriter_domain" "test" {
  name = "test.com"
  creation = "2024-01-01"
  expiration = "2025-01-01"
  force_delete = true
}

resource "ghostwriter_domain_checkout" "test" {
  project_id       = data.ghostwriter_project.testproject.id
  domain_id        = resource.ghostwriter_domain.test.id
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.test.id
  force_delete = true
}

resource "ghostwriter_cloud_server" "test" {
  server_provider_id = 1
  activity_type_id = data.ghostwriter_activity_type.test.id
  ip_address = "192.168.0.1"
  aux_address = ["192.168.0.1"]
  project_id = data.ghostwriter_project.testproject.id
  server_role_id = 1
  force_delete = true
}

resource "ghostwriter_domain_server" "test" {
  domain_checkout_id       = resource.ghostwriter_domain_checkout.test.id
  project_id        = data.ghostwriter_project.testproject.id
  cloud_server_id       = resource.ghostwriter_cloud_server.test.id
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "domain_checkout_id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "project_id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "cloud_server_id"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "subdomain", "*"),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "endpoint", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain_server.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_server.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
