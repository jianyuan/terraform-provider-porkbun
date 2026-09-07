package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwacctest"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDomainResource(t *testing.T) {
	t.Parallel()

	if !acctest.TestRunDomainRegistration {
		t.Skip("PORKBUN_RUN_DOMAIN_REGISTRATION is not set to 1")
	}

	rn := "porkbun_domain.test"
	domain := acctest.RandomDomain()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfig(domain, `
					max_cost = 1
				`),
				ExpectError: fwacctest.ExpectLiteralError(`above the 1 permitted by max_cost. Nothing was charged.`),
			},
			{
				Config: testAccDomainResourceConfig(domain, ``),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(domain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("cost"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("status"), knownvalue.StringExact("ACTIVE")),
				},
			},
			{
				Config:                               testAccDomainResourceConfig(domain, ``),
				ResourceName:                         rn,
				ImportState:                          true,
				ImportStateId:                        domain,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
				ImportStateVerifyIgnore:              []string{"cost", "order_id"},
			},
		},
	})
}

func testAccDomainResourceConfig(domain, extras string) string {
	return fmt.Sprintf(`
resource "porkbun_domain" "test" {
	domain = "%[1]s"
	%[2]s
}
`, domain, extras)
}
