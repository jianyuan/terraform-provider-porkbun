package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

func NewDomainNameserversResource() resource.Resource {
	return &DomainNameserversResource{}
}

var _ resource.Resource = &DomainNameserversResource{}

type DomainNameserversResource struct {
	baseResource
}

func (r *DomainNameserversResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_nameservers"
}

func (r *DomainNameserversResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = domainNameserversSchema().GetResource(ctx)
}

func (r *DomainNameserversResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainNameserversModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nameservers []string
	resp.Diagnostics.Append(data.Nameservers.ElementsAs(ctx, &nameservers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createHttpResp, err := r.client.DomainUpdateNsWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainUpdateNsJSONRequestBody{
			Ns: nameservers,
		},
	)

	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if !apiclient.IsOK(createHttpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(createHttpResp))
		return
	}

	readHttpResp, err := r.client.GetDomainNsWithResponse(
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
	} else if readHttpResp.JSON200.Ns == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no name servers")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *readHttpResp.JSON200.Ns)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainNameserversModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetDomainNsWithResponse(
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
	} else if httpResp.JSON200.Ns == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no name servers")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *httpResp.JSON200.Ns)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainNameserversModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nameservers []string
	resp.Diagnostics.Append(data.Nameservers.ElementsAs(ctx, &nameservers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateHttpResp, err := r.client.DomainUpdateNsWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainUpdateNsJSONRequestBody{
			Ns: nameservers,
		},
	)

	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientUpdateError(err))
		return
	} else if !apiclient.IsOK(updateHttpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientUpdateHTTPResponseError(updateHttpResp))
		return
	}

	readHttpResp, err := r.client.GetDomainNsWithResponse(
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
	} else if readHttpResp.JSON200.Ns == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no name servers")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *readHttpResp.JSON200.Ns)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Not Supported", "Deletion is not supported for this resource. Nameservers will be left unchanged.")
}
