package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestActivityTypeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "ghostwriter_activity_type" "test" {
  name = "Command and Control"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ghostwriter_domain.test", "name", "Command and Control"),
					resource.TestCheckResourceAttrSet("ghostwriter_domain.test", "id"),
				),
			},
		},
	})
}
