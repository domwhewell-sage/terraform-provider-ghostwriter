package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDomainCheckoutResource(t *testing.T) {
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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "project_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "domain_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "end_date", "2025-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "activity_type_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "note", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_checkout.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_checkout.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ghostwriter_domain_checkout.test",
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

resource "ghostwriter_domain" "updatedtest" {
  name = "updatedtest.com"
  creation = "2024-01-01"
  expiration = "2025-01-01"
  force_delete = true
}

resource "ghostwriter_domain_checkout" "test" {
  project_id       = 1
  domain_id        = resource.ghostwriter_domain.updatedtest.id
  start_date       = data.ghostwriter_project.testproject.start_date
  end_date         = data.ghostwriter_project.testproject.end_date
  activity_type_id = data.ghostwriter_activity_type.test.id
  note			 = "test note"
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "project_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "domain_id", "2"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "start_date", "2024-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "end_date", "2025-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "activity_type_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "note", "test note"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_checkout.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain_checkout.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
