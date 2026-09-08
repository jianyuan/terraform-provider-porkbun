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

func TestAccGlueRecordsDataSource(t *testing.T) {
	t.Parallel()

	rn := "data.porkbun_glue_records.test"
	subdomain := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueRecordsDataSourceConfig(subdomain, []string{"192.0.2.1"}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("hosts"), knownvalue.MapPartial(map[string]knownvalue.Check{
						fmt.Sprintf("%s.%s", subdomain, acctest.TestDomain): knownvalue.ObjectExact(map[string]knownvalue.Check{
							"ips": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("192.0.2.1"),
							}),
						}),
					})),
				},
			},
		},
	})
}

func testAccGlueRecordsDataSourceConfig(subdomain string, ips []string) string {
	return testAccGlueRecordResourceConfig(subdomain, ips) + `
data "porkbun_glue_records" "test" {
	domain = porkbun_glue_record.test.domain
}
`
}
