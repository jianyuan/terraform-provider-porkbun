package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func (m *DnsRecordResourceModel) Fill(ctx context.Context, record apiclient.DnsRecordsResponse_Records) (diags diag.Diagnostics) {
	diags.Append(m.DnsRecordModel.Fill(ctx, record)...)
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a DNS record for a domain.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain for the record being created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The record ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subdomain": schema.StringAttribute{
				MarkdownDescription: "The subdomain for the record being created, not including the domain itself. Omit to create a record on the root domain. Use * to create a wildcard record.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The full name of the record being created, including the subdomain and the domain itself.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of record being created. Valid types are: A, MX, CNAME, ALIAS, TXT, NS, AAAA, SRV, TLSA, CAA, HTTPS, SVCB.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(DnsRecordTypes...),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The answer content for the record. Please see the DNS management popup from the domain management console for proper formatting of each record type.",
				Required:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "The time to live in seconds for the record. The minimum and the default is 600 seconds.",
				Optional:            true,
				Computed:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority of the record for those that support it.",
				Optional:            true,
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				Computed: true,
			},
		},
	}
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
	}

	httpResp, err := r.client.DnsCreateWithResponse(
		ctx,
		data.Domain.ValueString(),
		body,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status == nil || *httpResp.JSON200.Status != "SUCCESS" {
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
		Name:    data.Name.ValueStringPointer(),
		Type:    apiclient.EditDnsRequestType(data.Type.ValueString()),
		Content: data.Content.ValueString(),
		Prio:    data.Priority.ValueInt64Pointer(),
		Ttl:     data.Ttl.ValueInt64Pointer(),
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
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
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
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
		diags.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if len(httpResp.JSON200.Records) != 1 {
		diags.AddError("Client error", fmt.Sprintf("Expected exactly one record, got %d", len(httpResp.JSON200.Records)))
		return
	}

	diags.Append(data.Fill(ctx, httpResp.JSON200.Records[0])...)
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
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
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
