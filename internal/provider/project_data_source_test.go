package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestProjectDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "ghostwriter_project" "test" {
  code_name = "TestProject"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ghostwriter_project.test", "code_name", "TestProject"),
					resource.TestCheckResourceAttrSet("data.ghostwriter_project.test", "id"),
				),
			},
		},
	})
}
