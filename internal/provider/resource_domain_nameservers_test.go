package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDomainNameserversResource(t *testing.T) {
	rn := "porkbun_domain_nameservers.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameserversResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
				},
			},
		},
	})
}

func testAccDomainNameserversResourceConfig() string {
	return fmt.Sprintf(`
resource "porkbun_domain_nameservers" "test" {
	domain = "%[1]s"

	nameservers = [
		"gabe.ns.cloudflare.com",
		"ivy.ns.cloudflare.com",
	]
}
`, acctest.TestDomain)
}
