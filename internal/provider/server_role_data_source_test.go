package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServerRoleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "ghostwriter_server_role" "test" {
  name = "Team Server / C2 Server"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ghostwriter_server_role.test", "name", "Team Server / C2 Server"),
					resource.TestCheckResourceAttrSet("data.ghostwriter_server_role.test", "id"),
				),
			},
		},
	})
}
