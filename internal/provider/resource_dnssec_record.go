package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

type DnssecRecordResourceModel struct {
	Domain          types.String `tfsdk:"domain"`
	Alg             types.String `tfsdk:"alg"`
	Digest          types.String `tfsdk:"digest"`
	DigestType      types.String `tfsdk:"digest_type"`
	KeyTag          types.String `tfsdk:"key_tag"`
	KeyDataAlgo     types.String `tfsdk:"key_data_algo"`
	KeyDataFlags    types.String `tfsdk:"key_data_flags"`
	KeyDataProtocol types.String `tfsdk:"key_data_protocol"`
	KeyDataPubKey   types.String `tfsdk:"key_data_public_key"`
	MaxSigLife      types.String `tfsdk:"max_sig_life"`
}

func (m *DnssecRecordResourceModel) Fill(ctx context.Context, data apiclient.DnsGetDnssecRecords200JSONResponseBody_Records) (diags diag.Diagnostics) {
	m.Alg = types.StringPointerValue(data.Alg)
	m.Digest = types.StringPointerValue(data.Digest)
	m.DigestType = types.StringPointerValue(data.DigestType)
	m.KeyTag = types.StringPointerValue(data.KeyTag)
	m.KeyDataPubKey = types.StringPointerValue(data.PubKey)
	return
}

func NewDnssecRecordResource() resource.Resource {
	return &DnssecRecordResource{}
}

var _ resource.Resource = &DnssecRecordResource{}
var _ resource.ResourceWithImportState = &DnssecRecordResource{}

type DnssecRecordResource struct {
	baseResource
}

func (r *DnssecRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_record"
}

func (r *DnssecRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a DNSSEC DS or key record at the registry. DNSSEC requirements vary by registry — `key_tag`, `alg`, `digest_type`, and `digest` are the minimum required fields. Key data fields are optional and will be omitted if not accepted by the registry.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain for the record being created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alg": schema.StringAttribute{
				MarkdownDescription: "DS Data algorithm number (e.g. 13 for ECDSA P-256 SHA-256).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"digest": schema.StringAttribute{
				MarkdownDescription: "Hex-encoded digest value.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"digest_type": schema.StringAttribute{
				MarkdownDescription: "Digest type number (e.g. 2 for SHA-256).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_tag": schema.StringAttribute{
				MarkdownDescription: "DNSSEC key tag.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_data_algo": schema.StringAttribute{
				MarkdownDescription: "Key data algorithm.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_data_flags": schema.StringAttribute{
				MarkdownDescription: "Key data flags (optional, used when submitting full key data).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_data_protocol": schema.StringAttribute{
				MarkdownDescription: "Key data protocol.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_data_public_key": schema.StringAttribute{
				MarkdownDescription: "Key data public key in base64.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_sig_life": schema.StringAttribute{
				MarkdownDescription: "Maximum signature lifetime in seconds (registry-specific).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DnssecRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DnssecRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := apiclient.DnsCreateDnssecRecordJSONRequestBody{
		Alg:             data.Alg.ValueString(),
		Digest:          data.Digest.ValueString(),
		DigestType:      data.DigestType.ValueString(),
		KeyDataAlgo:     data.KeyDataAlgo.ValueStringPointer(),
		KeyDataFlags:    data.KeyDataFlags.ValueStringPointer(),
		KeyDataProtocol: data.KeyDataProtocol.ValueStringPointer(),
		KeyDataPubKey:   data.KeyDataPubKey.ValueStringPointer(),
		KeyTag:          data.KeyTag.ValueString(),
		MaxSigLife:      data.MaxSigLife.ValueStringPointer(),
	}

	createHttpResp, err := r.client.DnsCreateDnssecRecordWithResponse(
		ctx,
		data.Domain.ValueString(),
		params,
	)

	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if createHttpResp.StatusCode() != http.StatusOK || createHttpResp.JSON200 == nil || createHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(createHttpResp))
		return
	}

	readHttpResp, err := r.client.DnsGetDnssecRecordsWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DnsGetDnssecRecordsJSONRequestBody{},
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if readHttpResp.StatusCode() != http.StatusOK || readHttpResp.JSON200 == nil || readHttpResp.JSON200.Status == nil || *readHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(readHttpResp))
		return
	} else if readHttpResp.JSON200.Records == nil {
		resp.Diagnostics.AddError("Client error", "Expected at least one DNSSEC record")
		return
	}

	record, ok := (*readHttpResp.JSON200.Records)[data.KeyTag.ValueString()]
	if !ok {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Expected record with key tag %s, but not found", data.KeyTag.ValueString()))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, record)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DnssecRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DnssecRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DnsGetDnssecRecordsWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DnsGetDnssecRecordsJSONRequestBody{},
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status == nil || *httpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if httpResp.JSON200.Records == nil {
		resp.Diagnostics.AddError("Client error", "Expected at least one DNSSEC record")
		return
	}

	record, ok := (*httpResp.JSON200.Records)[data.KeyTag.ValueString()]
	if !ok {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Expected record with key tag %s, but not found", data.KeyTag.ValueString()))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, record)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DnssecRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Not Supported", "Update is not supported for this resource.")
}

func (r *DnssecRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DnssecRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DnsDeleteDnssecRecordWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.KeyTag.ValueString(),
		apiclient.DnsDeleteDnssecRecordJSONRequestBody{},
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

func (r *DnssecRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected import ID in the format '<domain>_<key_tag>' (e.g. jiancodes.com_12345).",
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("key_tag"), parts[1])
}
