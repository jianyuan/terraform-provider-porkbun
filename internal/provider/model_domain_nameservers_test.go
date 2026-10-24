package provider

import (
	"context"
	"slices"
	"testing"

	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

func TestDomainNameserversModelFromAPI(t *testing.T) {
	ctx := context.Background()
	m := DomainNameserversModel{
		Nameservers: supertypes.NewSetValueOfSlice[string](ctx, nil),
	}

	diags := m.FromAPI(ctx, []string{
		"ns1.awsdns-00.com.",
		"ns2.awsdns-01.com",
		"ns1.awsdns-00.com",
		"ns3.awsdns-02.com.",
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	got, diags := m.Nameservers.Get(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	slices.Sort(got)
	want := []string{"ns1.awsdns-00.com", "ns2.awsdns-01.com", "ns3.awsdns-02.com"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
