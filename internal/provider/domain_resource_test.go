package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDomainResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ghostwriter_domain" "test" {
  name = "test.com"
  creation = "2024-01-01"
  expiration = "2025-01-01"
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "name", "test.com"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "registrar", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "creation", "2024-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "expiration", "2025-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "auto_renew", "false"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "note", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "burned_explanation", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "vt_permalink", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ghostwriter_domain.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"force_delete", "last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ghostwriter_domain" "test" {
  name = "updatedtest.com"
  registrar = "amazon"
  creation = "2024-01-01"
  expiration = "2025-01-01"
  auto_renew = true
  note = "test note"
  force_delete = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "name", "updatedtest.com"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "registrar", "amazon"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "creation", "2024-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "expiration", "2025-01-01"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "auto_renew", "true"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "note", "test note"),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "burned_explanation", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "vt_permalink", ""),
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "force_delete", "true"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
