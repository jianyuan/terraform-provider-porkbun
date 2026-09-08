package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
)

type DnsRecordResourceModel struct {
	Domain    types.String `tfsdk:"domain"`
	Subdomain types.String `tfsdk:"subdomain"`
	DnsRecordModel
}

func (m *DnsRecordResourceModel) FromAPI(ctx context.Context, record apiclient.DnsRecordsResponse_Records) (diags diag.Diagnostics) {
	diags.Append(m.DnsRecordModel.FromAPI(ctx, record)...)
	if diags.HasError() {
		return
	}

	fullName := strings.TrimSuffix(m.Name.ValueString(), ".")
	domain := strings.TrimSuffix(m.Domain.ValueString(), ".")
	if fullName == domain {
		m.Subdomain = types.StringNull()
	} else if suffix := "." + domain; strings.HasSuffix(fullName, suffix) {
		m.Subdomain = types.StringValue(strings.TrimSuffix(fullName, suffix))
	}
	return
}

func NewDnsRecordResource() resource.Resource {
	return &DnsRecordResource{}
}

var _ resource.Resource = &DnsRecordResource{}
var _ resource.ResourceWithImportState = &DnsRecordResource{}

type DnsRecordResource struct {
	baseResource
}

func (r *DnsRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DnsRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = dnsRecordSchema().GetResource(ctx)
}

func (r *DnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DnsRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.DnsCreateJSONRequestBody{
		Name:    data.Subdomain.ValueStringPointer(),
		Type:    apiclient.CreateDnsRequestType(data.Type.ValueString()),
		Content: data.Content.ValueString(),
		Prio:    data.Priority.ValueInt64Pointer(),
		Ttl:     data.Ttl.ValueInt64Pointer(),
		Notes:   data.Notes.ValueStringPointer(),
	}

	httpResp, err := r.client.DnsCreateWithResponse(
		ctx,
		data.Domain.ValueString(),
		body,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	data.Id = porkbuntypes.StringIDPointerValue(httpResp.JSON200.Id)

	resp.Diagnostics.Append(r.read(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DnsRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.read(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DnsRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.DnsEditJSONRequestBody{
		Name:    data.Subdomain.ValueStringPointer(),
		Type:    apiclient.EditDnsRequestType(data.Type.ValueString()),
		Content: data.Content.ValueString(),
		Prio:    data.Priority.ValueInt64Pointer(),
		Ttl:     data.Ttl.ValueInt64Pointer(),
		Notes:   data.Notes.ValueStringPointer(),
	}

	httpResp, err := r.client.DnsEditWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		body,
	)

	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientUpdateError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientUpdateHTTPResponseError(httpResp))
		return
	}

	resp.Diagnostics.Append(r.read(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DnsRecordResource) read(ctx context.Context, data *DnsRecordResourceModel) (diags diag.Diagnostics) {
	httpResp, err := r.client.GetDnsRecordByIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		nil,
	)
	if err != nil {
		diags.Append(fwdiag.NewClientReadError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		diags.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if len(httpResp.JSON200.Records) != 1 {
		diags.AddError("Client error", fmt.Sprintf("Expected exactly one record, got %d", len(httpResp.JSON200.Records)))
		return
	}

	diags.Append(data.FromAPI(ctx, httpResp.JSON200.Records[0])...)
	return
}

func (r *DnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DnsRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DnsDeleteWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		apiclient.DnsDeleteJSONRequestBody{},
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientDeleteError(err))
		return
	} else if httpResp.StatusCode() == http.StatusNotFound {
		return
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientDeleteHTTPResponseError(httpResp))
		return
	}
}

func (r *DnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected import ID in the format '<record_id>_<domain>_<type>' (e.g. 123456789_jiancodes.com_CNAME).",
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("id"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("domain"), parts[1])
	resp.State.SetAttribute(ctx, path.Root("type"), parts[2])
}
