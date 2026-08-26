package provider

import (
	"fmt"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
	"github.com/samber/lo"
)

func TestAccGlueRecordResource(t *testing.T) {
	rn := "porkbun_glue_record.test"
	subdomain := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueRecordResourceConfig(subdomain, []string{"192.0.2.1"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.StringExact(subdomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ips"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("192.0.2.1"),
					})),
				},
			},
			{
				Config: testAccGlueRecordResourceConfig(subdomain, []string{"192.0.2.1", "2001:db8::68"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ips"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("192.0.2.1"),
						knownvalue.StringExact("2001:db8::68"),
					})),
				},
			},
			{
				Config:                               testAccGlueRecordResourceConfig(subdomain, []string{"192.0.2.1", "2001:db8::68"}),
				ResourceName:                         rn,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGlueRecordImportStateIdFunc(rn),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "subdomain",
			},
		},
	})
}

func testAccGlueRecordImportStateIdFunc(rn string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return "", fmt.Errorf("not found: %s", rn)
		}
		return fmt.Sprintf("%s_%s", rs.Primary.Attributes["domain"], rs.Primary.Attributes["subdomain"]), nil
	}
}

func testAccGlueRecordResourceConfig(subdomain string, ips []string) string {
	ipList := strings.Join(lo.Map(ips, func(ip string, _ int) string {
		return fmt.Sprintf("\"%s\"", ip)
	}), ",")

	return fmt.Sprintf(`
resource "porkbun_glue_record" "test" {
	domain    = "%[1]s"
	subdomain = "%[2]s"
	ips       = [%[3]s]
}
`, acctest.TestDomain, subdomain, ipList)
}
