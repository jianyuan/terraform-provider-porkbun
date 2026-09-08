package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
)

type DomainAutoRenewSettingResourceModel struct {
	Domain    types.String `tfsdk:"domain"`
	AutoRenew types.Bool   `tfsdk:"auto_renew"`
}

func (m *DomainAutoRenewSettingResourceModel) FromAPI(ctx context.Context, autoRenew *porkbuntypes.FlexibleBool) (diags diag.Diagnostics) {
	m.AutoRenew = porkbuntypes.FlexibleBoolPointerValue(autoRenew)
	return
}

func NewDomainAutoRenewSettingResource() resource.Resource {
	return &DomainAutoRenewSettingResource{}
}

var _ resource.Resource = &DomainAutoRenewSettingResource{}
var _ resource.ResourceWithConfigure = &DomainAutoRenewSettingResource{}
var _ resource.ResourceWithImportState = &DomainAutoRenewSettingResource{}

type DomainAutoRenewSettingResource struct {
	baseResource
}

func (r *DomainAutoRenewSettingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_auto_renew_setting"
}

func (r *DomainAutoRenewSettingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Update the auto-renew setting for one domain.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain to update the auto-renew setting for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_renew": schema.BoolAttribute{
				MarkdownDescription: "Auto-renew status to set.",
				Required:            true,
			},
		},
	}
}

func (r *DomainAutoRenewSettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainAutoRenewSettingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.DomainUpdateAutoRenewJSONRequestBody{}
	if data.AutoRenew.ValueBool() {
		body.Status = apiclient.UpdateAutoRenewRequestStatusOn
	} else {
		body.Status = apiclient.UpdateAutoRenewRequestStatusOff
	}

	updateHttpResp, err := r.client.DomainUpdateAutoRenewWithResponse(
		ctx,
		data.Domain.ValueString(),
		body,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if !apiclient.IsOK(updateHttpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(updateHttpResp))
		return
	}

	readHttpResp, err := r.client.GetDomainWithResponse(
		ctx,
		data.Domain.ValueString(),
		nil,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if !apiclient.IsOK(readHttpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(readHttpResp))
		return
	} else if readHttpResp.JSON200.Domain == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no domain")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, readHttpResp.JSON200.Domain.AutoRenew)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainAutoRenewSettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainAutoRenewSettingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetDomainWithResponse(
		ctx,
		data.Domain.ValueString(),
		nil,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if httpResp.JSON200.Domain == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no domain")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, httpResp.JSON200.Domain.AutoRenew)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainAutoRenewSettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainAutoRenewSettingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.DomainUpdateAutoRenewJSONRequestBody{}
	if data.AutoRenew.ValueBool() {
		body.Status = apiclient.UpdateAutoRenewRequestStatusOn
	} else {
		body.Status = apiclient.UpdateAutoRenewRequestStatusOff
	}

	updateHttpResp, err := r.client.DomainUpdateAutoRenewWithResponse(
		ctx,
		data.Domain.ValueString(),
		body,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if !apiclient.IsOK(updateHttpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(updateHttpResp))
		return
	}

	readHttpResp, err := r.client.GetDomainWithResponse(
		ctx,
		data.Domain.ValueString(),
		nil,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if !apiclient.IsOK(readHttpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(readHttpResp))
		return
	} else if readHttpResp.JSON200.Domain == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no domain")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, readHttpResp.JSON200.Domain.AutoRenew)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainAutoRenewSettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Not Supported", "Delete is not supported for this resource.")
}

func (r *DomainAutoRenewSettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}
