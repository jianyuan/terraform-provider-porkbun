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

func TestAccWebhookResource(t *testing.T) {
	t.Parallel()

	rn := "porkbun_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "porkbun_webhook" "test" {
	url = "https://example.com/porkbun/webhook"
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("url"), knownvalue.StringExact("https://example.com/porkbun/webhook")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("events"), knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("*")})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("secret"), knownvalue.NotNull()),
				},
			},
			{
				Config: testAccWebhookResourceConfig("https://example.org/hook", false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("url"), knownvalue.StringExact("https://example.org/hook")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("events"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("domain.registered"),
						knownvalue.StringExact("dns.*"),
					})),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("enabled"), knownvalue.Bool(false)),
				},
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWebhookResource_createDisabled(t *testing.T) {
	t.Parallel()

	rn := "porkbun_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookResourceConfig("https://example.com/porkbun/webhook", false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccWebhookResourceConfig(url string, enabled bool) string {
	return fmt.Sprintf(`
resource "porkbun_webhook" "test" {
	url     = %q
	events  = ["domain.registered", "dns.*"]
	enabled = %t
}
`, url, enabled)
}
