package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
)

type DomainModel struct {
	Domain       types.String `tfsdk:"domain"`
	Status       types.String `tfsdk:"status"`
	Tld          types.String `tfsdk:"tld"`
	CreateDate   types.String `tfsdk:"create_date"`
	ExpireDate   types.String `tfsdk:"expire_date"`
	SecurityLock types.Bool   `tfsdk:"security_lock"`
	WhoisPrivacy types.Bool   `tfsdk:"whois_privacy"`
	AutoRenew    types.Bool   `tfsdk:"auto_renew"`
	ApiAccess    types.Bool   `tfsdk:"api_access"`
	NotLocal     types.Bool   `tfsdk:"not_local"`
}

func (m *DomainModel) FromAPI(ctx context.Context, domain apiclient.Domain) (diags diag.Diagnostics) {
	m.Domain = types.StringPointerValue(domain.Domain)
	m.Status = types.StringPointerValue(domain.Status)
	m.Tld = types.StringPointerValue(domain.Tld)
	m.CreateDate = types.StringPointerValue(domain.CreateDate)
	m.ExpireDate = types.StringPointerValue(domain.ExpireDate)
	m.SecurityLock = porkbuntypes.FlexibleBoolPointerValue(domain.SecurityLock)
	m.WhoisPrivacy = porkbuntypes.FlexibleBoolPointerValue(domain.WhoisPrivacy)
	m.AutoRenew = porkbuntypes.FlexibleBoolPointerValue(domain.AutoRenew)
	m.ApiAccess = porkbuntypes.FlexibleBoolPointerValue(domain.ApiAccess)
	m.NotLocal = porkbuntypes.FlexibleBoolPointerValue(domain.NotLocal)
	return
}
