package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServerProviderDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "ghostwriter_server_provicer" "test" {
  name = "Amazon Web Services"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ghostwriter_server_provicer.test", "name", "Amazon Web Services"),
					resource.TestCheckResourceAttrSet("data.ghostwriter_server_provicer.test", "id"),
				),
			},
		},
	})
}
