package provider

import (
	"context"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwtypes"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

type URLForwardResourceModel struct {
	Domain types.String `tfsdk:"domain"`
	URLForwardModel
}

func NewURLForwardResource() resource.Resource {
	return &URLForwardResource{}
}

var _ resource.Resource = &URLForwardResource{}
var _ resource.ResourceWithConfigure = &URLForwardResource{}
var _ resource.ResourceWithImportState = &URLForwardResource{}

type URLForwardResource struct {
	baseResource
}

func (r *URLForwardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_url_forward"
}

func (r *URLForwardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = urlForwardSchema().GetResource(ctx)
}

func (r *URLForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data URLForwardResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := apiclient.DomainAddUrlForwardJSONRequestBody{
		Location:  data.Location.ValueString(),
		Subdomain: data.Subdomain.ValueStringPointer(),
		Type:      apiclient.AddUrlForwardRequestType(data.Type.ValueString()),
	}
	if data.IncludePath.ValueBool() {
		params.IncludePath = apiclient.AddUrlForwardRequestIncludePathYes
	} else {
		params.IncludePath = apiclient.AddUrlForwardRequestIncludePathNo
	}
	if fwtypes.IsKnown(data.RedirectType) {
		params.RedirectType = new(apiclient.AddUrlForwardRequestRedirectType(data.RedirectType.ValueString()))
	}
	if data.Wildcard.ValueBool() {
		params.Wildcard = apiclient.AddUrlForwardRequestWildcardYes
	} else {
		params.Wildcard = apiclient.AddUrlForwardRequestWildcardNo
	}

	httpResp, err := r.client.DomainAddUrlForwardWithResponse(
		ctx,
		data.Domain.ValueString(),
		params,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	// NOTE: The create response does not contain the ID, so we need to read it back
	// and set the ID on the data

	resp.Diagnostics.Append(r.read(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *URLForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data URLForwardResourceModel

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

func (r *URLForwardResource) read(ctx context.Context, data *URLForwardResourceModel) (diags diag.Diagnostics) {
	httpResp, err := r.client.GetDomainUrlForwardingWithResponse(ctx, data.Domain.ValueString(), nil)
	if err != nil {
		diags.Append(fwdiag.NewClientReadError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		diags.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if httpResp.JSON200.Forwards == nil {
		diags.AddError("Client error", "Unable to read, got no URL forwards")
		return
	}

	dSubdomain := data.Subdomain.ValueString()
	dLocation := data.Location.ValueString()
	dType := apiclient.GetUrlForwardingResponseForwardsType(data.Type.ValueString())
	dRedirectType := apiclient.GetUrlForwardingResponseForwardsRedirectType(data.RedirectType.ValueString())
	dIncludePath := apiclient.GetUrlForwardingResponseForwardsIncludePathNo
	if data.IncludePath.ValueBool() {
		dIncludePath = apiclient.GetUrlForwardingResponseForwardsIncludePathYes
	}
	dWildcard := apiclient.GetUrlForwardingResponseForwardsWildcardNo
	if data.Wildcard.ValueBool() {
		dWildcard = apiclient.GetUrlForwardingResponseForwardsWildcardYes
	}

	for _, forward := range httpResp.JSON200.Forwards {
		if forward.Id == nil {
			continue
		}

		if fwtypes.IsKnown(data.Id) {
			if *forward.Id != data.Id.ValueString() {
				continue
			}
		} else {
			if forward.Subdomain == nil || *forward.Subdomain != dSubdomain {
				continue
			}
			if forward.Location == nil || *forward.Location != dLocation {
				continue
			}
			if forward.Type == nil || *forward.Type != dType {
				continue
			}
			if forward.RedirectType == nil || (dRedirectType != "" && *forward.RedirectType != dRedirectType) {
				continue
			}
			if forward.IncludePath == nil || *forward.IncludePath != dIncludePath {
				continue
			}
			if forward.Wildcard == nil || *forward.Wildcard != dWildcard {
				continue
			}
		}

		diags.Append(data.FromAPI(ctx, forward)...)
		return
	}

	diags.AddError("Client error", "No matching URL forward found")
	return
}

func (r *URLForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Not Supported", "Update is not supported for this resource.")
}

func (r *URLForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data URLForwardResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DomainDeleteUrlForwardWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		apiclient.DomainDeleteUrlForwardJSONRequestBody{},
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

func (r *URLForwardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected import ID in the format '<url_forward_id>_<domain>' (e.g. 123456789_jiancodes.com).",
		)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("domain"), parts[1])
}
