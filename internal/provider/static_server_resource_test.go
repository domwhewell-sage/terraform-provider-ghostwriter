package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestStaticServerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ghostwriter_static_server" "test" {
  name = "hostname"
  server_provider_id = 1
  ip_address = "192.168.0.1"
  note = "Test Note"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "name", "hostname"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "server_provider_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "server_status_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "ip_address", "192.168.0.1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "note", "Test Note"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ghostwriter_static_server.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ghostwriter_static_server" "test" {
  name = "New Hostname"
  server_provider_id = 1
  ip_address = "192.168.0.2"
  note = "Test updated note"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "name", "New Hostname"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "server_provider_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "server_status_id", "1"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "ip_address", "192.168.0.2"),
					resource.TestCheckResourceAttr("ghostwriter_static_server.test", "note", "Test updated note"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server.test", "id"),
					resource.TestCheckResourceAttrSet("ghostwriter_static_server.test", "last_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
