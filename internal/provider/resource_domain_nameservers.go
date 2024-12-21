package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "Update the name servers for your domain.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nameservers": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
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

	createHttpResp, err := r.client.DomainUpdateNameServersWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainUpdateNameServersJSONRequestBody{
			Apikey:       r.apiKey,
			Secretapikey: r.secretKey,
			Ns:           nameservers,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create, got error: %s", err))
		return
	} else if createHttpResp.StatusCode() != http.StatusOK || createHttpResp.JSON200 == nil || createHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create, got status code %d: %s", createHttpResp.StatusCode(), string(createHttpResp.Body)))
		return
	}

	readHttpResp, err := r.client.DomainGetNameServersWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainGetNameServersJSONRequestBody{
			Apikey:       r.apiKey,
			Secretapikey: r.secretKey,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
		return
	} else if readHttpResp.StatusCode() != http.StatusOK || readHttpResp.JSON200 == nil || readHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code %d: %s", readHttpResp.StatusCode(), string(readHttpResp.Body)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, readHttpResp.JSON200.Ns)...)
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

	httpResp, err := r.client.DomainGetNameServersWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainGetNameServersJSONRequestBody{
			Apikey:       r.apiKey,
			Secretapikey: r.secretKey,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, httpResp.JSON200.Ns)...)
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

	updateHttpResp, err := r.client.DomainUpdateNameServersWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainUpdateNameServersJSONRequestBody{
			Apikey:       r.apiKey,
			Secretapikey: r.secretKey,
			Ns:           nameservers,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update, got error: %s", err))
		return
	} else if updateHttpResp.StatusCode() != http.StatusOK || updateHttpResp.JSON200 == nil || updateHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update, got status code %d: %s", updateHttpResp.StatusCode(), string(updateHttpResp.Body)))
		return
	}

	readHttpResp, err := r.client.DomainGetNameServersWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainGetNameServersJSONRequestBody{
			Apikey:       r.apiKey,
			Secretapikey: r.secretKey,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
		return
	} else if readHttpResp.StatusCode() != http.StatusOK || readHttpResp.JSON200 == nil || readHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code %d: %s", readHttpResp.StatusCode(), string(readHttpResp.Body)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, readHttpResp.JSON200.Ns)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainNameserversResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Not Supported", "Deletion is not supported for this resource. Nameservers will be left unchanged.")
}
