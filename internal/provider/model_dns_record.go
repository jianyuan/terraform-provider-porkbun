package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
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

func (m *DnsRecordModel) Fill(ctx context.Context, record apiclient.DnsRecord) (diags diag.Diagnostics) {
	m.Id = types.StringValue(record.Id)
	m.Name = types.StringValue(record.Name)
	m.Type = types.StringValue(record.Type)
	m.Content = types.StringValue(record.Content)

	if v, err := strconv.ParseInt(record.Ttl, 10, 64); err == nil {
		m.Ttl = types.Int64Value(v)
	} else {
		diags.AddError(fmt.Sprintf("failed to parse TTL: %s", err), fmt.Sprintf("failed to parse TTL: %s", err))
		return
	}

	if record.Prio == nil {
		m.Priority = types.Int64Null()
	} else if v, err := strconv.ParseInt(*record.Prio, 10, 64); err == nil {
		m.Priority = types.Int64Value(v)
	} else {
		diags.AddError(fmt.Sprintf("failed to parse PRIO: %s", err), fmt.Sprintf("failed to parse PRIO: %s", err))
		return
	}

	m.Notes = types.StringPointerValue(record.Notes)
	return
}
