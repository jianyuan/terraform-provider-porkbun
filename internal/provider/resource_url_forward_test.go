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

func TestAccURLForwardResource(t *testing.T) {
	t.Parallel()

	rn := "porkbun_url_forward.test"
	subdomain := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Temporary
			{
				Config: testAccURLForwardResourceConfig(`
					type = "temporary"
					wildcard = false
					include_path = false
					location = "https://www.jiancodes.com"
				`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("temporary")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("wildcard"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("include_path"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("location"), knownvalue.StringExact("https://www.jiancodes.com")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("redirect_type"), knownvalue.StringExact("302")),
				},
			},
			// Permanent
			{
				Config: testAccURLForwardResourceConfig(`
					type = "permanent"
					wildcard = false
					include_path = false
					location = "https://www.jiancodes.com"
				`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("permanent")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("wildcard"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("include_path"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("location"), knownvalue.StringExact("https://www.jiancodes.com")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("redirect_type"), knownvalue.StringExact("301")),
				},
			},
			// Masked
			{
				Config: testAccURLForwardResourceConfig(`
					type = "masked"
					wildcard = false
					include_path = false
					location = "https://www.jiancodes.com"
				`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("masked")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("wildcard"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("include_path"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("location"), knownvalue.StringExact("https://www.jiancodes.com")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("redirect_type"), knownvalue.StringExact("masked")),
				},
			},
			// Import without subdomain
			{
				Config: testAccURLForwardResourceConfig(`
					type = "masked"
					wildcard = false
					include_path = false
					location = "https://www.jiancodes.com"
				`),
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccURLForwardImportStateIdFunc(rn),
				ImportStateVerify: true,
			},
			{
				Config: testAccURLForwardResourceConfig(fmt.Sprintf(`
					subdomain = "%[1]s"
					type = "temporary"
					wildcard = true
					include_path = true
					location = "https://www.jiancodes.com"
					redirect_type = "307"
				`, subdomain)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.StringExact(subdomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("temporary")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("wildcard"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("include_path"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("location"), knownvalue.StringExact("https://www.jiancodes.com")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("redirect_type"), knownvalue.StringExact("307")),
				},
			},
			// Import with subdomain
			{
				Config: testAccURLForwardResourceConfig(fmt.Sprintf(`
					subdomain = "%[1]s"
					type = "temporary"
					wildcard = true
					include_path = true
					location = "https://www.jiancodes.com"
					redirect_type = "307"
				`, subdomain)),
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccURLForwardImportStateIdFunc(rn),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccURLForwardImportStateIdFunc(rn string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return "", fmt.Errorf("not found: %s", rn)
		}
		return fmt.Sprintf("%s_%s", rs.Primary.ID, rs.Primary.Attributes["domain"]), nil
	}
}

func testAccURLForwardResourceConfig(extras string) string {
	return fmt.Sprintf(`
resource "porkbun_url_forward" "test" {
	domain = "%[1]s"
	%[2]s
}
`, acctest.TestDomain, extras)
}
