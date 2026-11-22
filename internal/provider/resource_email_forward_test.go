package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

// NOTE: Adding a forward changes the MX and SPF records at the zone apex of
// the test domain, and deleting the last forward removes them again. The test
// domain must not carry other apex MX records; confirm_dns_changes is left at
// its default so that the test fails rather than deleting them.
func TestAccEmailForwardResource(t *testing.T) {
	t.Parallel()

	if !acctest.TestRunEmailForward {
		t.Skip("PORKBUN_RUN_EMAIL_FORWARD is not set to 1; the sandbox does not support /email endpoints")
	}

	rn := "porkbun_email_forward.test"
	address := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailForwardResourceConfig(address),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("address"), knownvalue.StringExact(address)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("forward_to"), knownvalue.StringExact("me@example.org")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("confirm_dns_changes"), knownvalue.Bool(false)),
				},
			},
			// The full address is the same forward: no replacement, no diff.
			{
				Config:   testAccEmailForwardResourceConfig(address + "@" + acctest.TestDomain),
				PlanOnly: false,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccEmailForwardImportStateIdFunc(rn),
				ImportStateVerify: true,
				// State keeps the configured spelling (full address); import yields the local part.
				ImportStateVerifyIgnore: []string{"address"},
			},
		},
	})
}

func testAccEmailForwardImportStateIdFunc(rn string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return "", fmt.Errorf("not found: %s", rn)
		}
		a := rs.Primary.Attributes
		return fmt.Sprintf("%s/%s/%s", a["domain"], a["address"], a["forward_to"]), nil
	}
}

func testAccEmailForwardResourceConfig(address string) string {
	return fmt.Sprintf(`
resource "porkbun_email_forward" "test" {
	domain     = %q
	address    = %q
	forward_to = "me@example.org"
}
`, acctest.TestDomain, address)
}
