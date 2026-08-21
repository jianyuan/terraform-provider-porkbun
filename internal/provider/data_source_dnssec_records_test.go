package provider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDnssecRecordsDataSource(t *testing.T) {
	rn := "data.porkbun_dnssec_records.test"

	flags := uint16(257) // 257 = KSK (Key Signing Key)
	protocol := uint8(3) // 3 = DNSSEC

	// KEY: ECDSA Curve P-256 (DNSSEC Algorithm 13)
	ecdsaPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %s", err)
	}
	ecdsaPubKeyBytes, err := ecdsaPrivKey.PublicKey.Bytes()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %s", err)
	}
	dsData1 := generateDSData(acctest.TestDomain, flags, protocol, uint8(13), ecdsaPubKeyBytes)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnssecRecordsDataSourceConfig(dsData1.Alg, dsData1.Digest, dsData1.DigestType, dsData1.KeyTag),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("records"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"key_tag":     knownvalue.StringExact(dsData1.KeyTag),
							"alg":         knownvalue.StringExact(dsData1.Alg),
							"digest_type": knownvalue.StringExact(dsData1.DigestType),
							"digest":      knownvalue.StringExact(dsData1.Digest),
						}),
					})),
				},
			},
		},
	})
}

func testAccDnssecRecordsDataSourceConfig(alg, digest, digestType, keyTag string) string {
	return testAccDnssecRecordResourceConfig(alg, digest, digestType, keyTag) + `
data "porkbun_dnssec_records" "test" {
	domain = porkbun_dnssec_record.test.domain
}
`
}
