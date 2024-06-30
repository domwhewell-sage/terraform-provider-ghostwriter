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
resource "ghostwriter_domain_checkout" "test" {
  project_id       = 1
  domain_id        = 1
  start_date       = "2024-01-01"
  end_date         = "2025-01-01"
  activity_type_id = 1
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
resource "ghostwriter_domain_checkout" "test" {
  project_id       = 1
  domain_id        = 1
  start_date       = "2024-01-01"
  end_date         = "2025-01-01"
  activity_type_id = 1
  note			 = "test note"
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "project_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_domain_checkout.test", "domain_id", "1"),
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
