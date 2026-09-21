package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccURLForwardsDataSource(t *testing.T) {
	t.Parallel()

	rn := "data.porkbun_url_forwards.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccURLForwardsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("url_forwards"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":            knownvalue.NotNull(),
							"subdomain":     knownvalue.Null(),
							"type":          knownvalue.StringExact("temporary"),
							"wildcard":      knownvalue.Bool(false),
							"include_path":  knownvalue.Bool(false),
							"location":      knownvalue.StringExact("https://www.jiancodes.com"),
							"redirect_type": knownvalue.StringExact("302"),
						}),
					})),
				},
			},
		},
	})
}

func testAccURLForwardsDataSourceConfig() string {
	return testAccURLForwardResourceConfig(`
	type = "temporary"
	wildcard = false
	include_path = false
	location = "https://www.jiancodes.com"
`) + `
data "porkbun_url_forwards" "test" {
	domain = porkbun_url_forward.test.domain
}
`
}
