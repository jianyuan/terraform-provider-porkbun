package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
)

type DnsRecordModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Content  types.String `tfsdk:"content"`
	Ttl      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
	Notes    types.String `tfsdk:"notes"`
}

func (m *DnsRecordModel) Fill(ctx context.Context, record apiclient.DnsRecordsResponse_Records) (diags diag.Diagnostics) {
	m.Id = types.StringPointerValue(record.Id)
	m.Name = types.StringPointerValue(record.Name)
	m.Type = types.StringPointerValue(record.Type)
	m.Content = types.StringPointerValue(record.Content)
	m.Ttl = porkbuntypes.FlexibleInt64PointerValue(record.Ttl)
	m.Priority = porkbuntypes.FlexibleInt64PointerValue(record.Prio)
	m.Notes = types.StringPointerValue(record.Notes)
	return
}
