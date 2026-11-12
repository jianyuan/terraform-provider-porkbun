package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccURLForwardsDataSource(t *testing.T) {
	t.Parallel()

	rn := "data.porkbun_url_forwards.test"
	// Use a unique subdomain so this test does not collide with other tests that
	// create forwards on the shared test domain in parallel.
	subdomain := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccURLForwardsDataSourceConfig(subdomain),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("url_forwards"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":            knownvalue.NotNull(),
							"subdomain":     knownvalue.StringExact(subdomain),
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

func testAccURLForwardsDataSourceConfig(subdomain string) string {
	return testAccURLForwardResourceConfig(fmt.Sprintf(`
	subdomain = "%s"
	type = "temporary"
	wildcard = false
	include_path = false
	location = "https://www.jiancodes.com"
`, subdomain)) + `
data "porkbun_url_forwards" "test" {
	domain = porkbun_url_forward.test.domain
}
`
}
