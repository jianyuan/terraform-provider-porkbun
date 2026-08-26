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
	"github.com/jianyuan/terraform-provider-porkbun/internal/tfutils"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type GlueRecordResourceModel struct {
	Domain    types.String                  `tfsdk:"domain"`
	Subdomain types.String                  `tfsdk:"subdomain"`
	Ips       supertypes.SetValueOf[string] `tfsdk:"ips"`
}

func (m *GlueRecordResourceModel) Fill(ctx context.Context, records []string) (diags diag.Diagnostics) {
	diags.Append(m.Ips.Set(ctx, records)...)
	return
}

func NewGlueRecordResource() resource.Resource {
	return &GlueRecordResource{}
}

var _ resource.Resource = &GlueRecordResource{}
var _ resource.ResourceWithConfigure = &GlueRecordResource{}
var _ resource.ResourceWithImportState = &GlueRecordResource{}

type GlueRecordResource struct {
	baseResource
}

func (r *GlueRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_glue_record"
}

func (r *GlueRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a glue record (host object) for a nameserver hostname under a domain. Use this resource when hosting a nameserver at a subdomain of the domain itself (e.g. `ns1.example.com`).",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain for the record being created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subdomain": schema.StringAttribute{
				MarkdownDescription: "The subdomain portion only (e.g. 'ns1' for ns1.example.com).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ips": schema.SetAttribute{
				MarkdownDescription: "The IP addresses (IPv4 and/or IPv6) to associate with the host record.",
				Required:            true,
				CustomType:          supertypes.NewSetTypeOf[string](ctx),
			},
		},
	}
}

func (r *GlueRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GlueRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.DomainCreateGlueJSONRequestBody{}
	body.Ips = tfutils.MergeDiagnostics(data.Ips.Get(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DomainCreateGlueWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Subdomain.ValueString(),
		body,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	resp.Diagnostics.Append(r.read(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GlueRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GlueRecordResourceModel

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

func (r *GlueRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GlueRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.DomainUpdateGlueJSONRequestBody{}
	body.Ips = tfutils.MergeDiagnostics(data.Ips.Get(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DomainUpdateGlueWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Subdomain.ValueString(),
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

func (r *GlueRecordResource) read(ctx context.Context, data *GlueRecordResourceModel) (diags diag.Diagnostics) {
	httpResp, err := r.client.GetDomainGlueWithResponse(
		ctx,
		data.Domain.ValueString(),
		nil,
	)
	if err != nil {
		diags.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status == nil || *httpResp.JSON200.Status != "SUCCESS" {
		diags.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if httpResp.JSON200.Hosts == nil {
		diags.AddError("Client error", "Unable to read, got no glue records")
		return
	}

	for _, host := range *httpResp.JSON200.Hosts {
		if len(host) != 2 {
			diags.AddError("Client error", "Unable to read, got unexpected glue record format")
			return
		}

		hostname, err := host[0].AsGetDomainGlue200JSONResponseBodyHosts0()
		if err != nil {
			diags.AddError("Client error", "Unable to read, got unexpected glue record format")
			return
		}

		if hostname != fmt.Sprintf("%s.%s", data.Subdomain.ValueString(), data.Domain.ValueString()) {
			continue
		}

		if ipAddresses, err := host[1].AsGetDomainGlue200JSONResponseBodyHosts1(); err == nil {
			diags.Append(data.Fill(ctx, ipAddresses)...)
			return
		} else if ipAddresses, err := host[1].AsGetDomainGlue200JSONResponseBodyHosts2(); err == nil {
			var ips []string
			if ipAddresses.V4 != nil {
				ips = append(ips, *ipAddresses.V4...)
			}
			if ipAddresses.V6 != nil {
				ips = append(ips, *ipAddresses.V6...)
			}
			diags.Append(data.Fill(ctx, ips)...)
			return
		} else {
			diags.AddError("Client error", "Unable to read, got unexpected glue record format")
			return
		}
	}

	diags.AddError("Client error", "No matching glue record found")
	return
}

func (r *GlueRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GlueRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DomainDeleteGlueWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Subdomain.ValueString(),
		apiclient.DomainDeleteGlueJSONRequestBody{},
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

func (r *GlueRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected import ID in the format '<domain>_<subdomain>' (e.g. jiancodes.com_ns1).",
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("subdomain"), parts[1])
}
