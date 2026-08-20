package provider

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDnssecRecordResource(t *testing.T) {
	rn := "porkbun_dnssec_record.test"

	flags := uint16(257) // 257 = KSK (Key Signing Key)
	protocol := uint8(3) // 3 = DNSSEC

	// KEY 1: ECDSA Curve P-256 (DNSSEC Algorithm 13)
	ecdsaPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %s", err)
	}
	ecdsaPubKeyBytes, err := ecdsaPrivKey.PublicKey.Bytes()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %s", err)
	}
	dsData1 := generateDSData(acctest.TestDomain, flags, protocol, uint8(13), ecdsaPubKeyBytes)

	// KEY 2: Ed25519 (DNSSEC Algorithm 15)
	ed25519PubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Ed25519 key: %s", err)
	}
	ed25519PubKeyBytes := []byte(ed25519PubKey)
	dsData2 := generateDSData(acctest.TestDomain, flags, protocol, uint8(15), ed25519PubKeyBytes)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnssecRecordResourceConfig(dsData1.Alg, dsData1.Digest, dsData1.DigestType, dsData1.KeyTag, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("alg"), knownvalue.StringExact(dsData1.Alg)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("digest"), knownvalue.StringExact(dsData1.Digest)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("digest_type"), knownvalue.StringExact(dsData1.DigestType)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("key_tag"), knownvalue.StringExact(dsData1.KeyTag)),
				},
			},
			{
				Config:                               testAccDnssecRecordResourceConfig(dsData1.Alg, dsData1.Digest, dsData1.DigestType, dsData1.KeyTag, ""),
				ResourceName:                         rn,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccDnssecRecordImportStateIdFunc(rn),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key_tag",
			},
			{
				Config: testAccDnssecRecordResourceConfig(dsData2.Alg, dsData2.Digest, dsData2.DigestType, dsData2.KeyTag, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("alg"), knownvalue.StringExact(dsData2.Alg)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("digest"), knownvalue.StringExact(dsData2.Digest)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("digest_type"), knownvalue.StringExact(dsData2.DigestType)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("key_tag"), knownvalue.StringExact(dsData2.KeyTag)),
				},
			},
		},
	})
}

func testAccDnssecRecordImportStateIdFunc(rn string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return "", fmt.Errorf("not found: %s", rn)
		}
		return fmt.Sprintf("%s_%s", rs.Primary.Attributes["domain"], rs.Primary.Attributes["key_tag"]), nil
	}
}

func testAccDnssecRecordResourceConfig(alg, digest, digestType, keyTag, extras string) string {
	return fmt.Sprintf(`
resource "porkbun_dnssec_record" "test" {
	domain = "%[1]s"
	alg = "%[2]s"
	digest = "%[3]s"
	digest_type = "%[4]s"
	key_tag = "%[5]s"
	%[6]s
}
`, acctest.TestDomain, alg, digest, digestType, keyTag, extras)
}

type dsData struct {
	Alg           string `json:"alg"`
	Digest        string `json:"digest"`
	DigestType    string `json:"digestType"`
	KeyTag        string `json:"keyTag"`
	KeyDataAlgo   string `json:"keyDataAlgo,omitempty"`
	KeyDataFlags  string `json:"keyDataFlags,omitempty"`
	KeyDataProto  string `json:"keyDataProtocol,omitempty"`
	KeyDataPubKey string `json:"keyDataPubKey,omitempty"`
}

func generateDSData(domain string, flags uint16, protocol uint8, algorithm uint8, pubKeyBytes []byte) dsData {
	// 1. Build DNSKEY RDATA: Flags (2 bytes) | Protocol (1 byte) | Algorithm (1 byte) | Public Key
	dnskeyRdata := make([]byte, 4+len(pubKeyBytes))
	binary.BigEndian.PutUint16(dnskeyRdata[0:2], flags)
	dnskeyRdata[2] = protocol
	dnskeyRdata[3] = algorithm
	copy(dnskeyRdata[4:], pubKeyBytes)

	// 2. Build Wire Format Name (e.g. "example.com." -> \x07example\x03com\x00)
	var nameWire []byte
	trimmedDomain := strings.TrimSuffix(domain, ".")
	if trimmedDomain != "" {
		for _, part := range strings.Split(trimmedDomain, ".") {
			nameWire = append(nameWire, byte(len(part)))
			nameWire = append(nameWire, []byte(part)...)
		}
	}
	nameWire = append(nameWire, 0x00) // Root label terminator

	// 3. Compute Key Tag
	var ac uint32
	for i, b := range dnskeyRdata {
		if i&1 == 1 {
			ac += uint32(b)
		} else {
			ac += uint32(b) << 8
		}
	}
	ac += (ac >> 16) & 0xFFFF
	keyTag := uint16(ac & 0xFFFF)

	// 4. Digest input is Domain Wire Format + DNSKEY RDATA
	dsDigestInput := append(nameWire, dnskeyRdata...)
	digest := sha256.Sum256(dsDigestInput)

	// 5. Construct and return populated struct
	const sha256DigestType = "2"

	return dsData{
		Alg:        fmt.Sprintf("%d", algorithm),
		Digest:     strings.ToUpper(hex.EncodeToString(digest[:])),
		DigestType: sha256DigestType,
		KeyTag:     fmt.Sprintf("%d", keyTag),

		// Optional KeyData fields populated for completeness
		KeyDataAlgo:  fmt.Sprintf("%d", algorithm),
		KeyDataFlags: fmt.Sprintf("%d", flags),
		KeyDataProto: fmt.Sprintf("%d", protocol),
	}
}
