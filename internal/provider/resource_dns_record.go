package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/go-utils/ptr"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

type DnsRecordResourceModel struct {
	Domain    types.String `tfsdk:"domain"`
	Subdomain types.String `tfsdk:"subdomain"`
	DnsRecordModel
}

func NewDnsRecordResource() resource.Resource {
	return &DnsRecordResource{}
}

var _ resource.Resource = &DnsRecordResource{}

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

	params := apiclient.DnsCreateRecordJSONRequestBody{
		Apikey:       r.apiKey,
		Secretapikey: r.secretKey,
		Name:         data.Subdomain.ValueStringPointer(),
		Type:         data.Type.ValueString(),
		Content:      data.Content.ValueString(),
	}
	if !data.Ttl.IsNull() {
		params.Ttl = ptr.Ptr(strconv.FormatInt(data.Ttl.ValueInt64(), 10))
	}
	if !data.Priority.IsNull() {
		params.Prio = ptr.Ptr(strconv.FormatInt(data.Priority.ValueInt64(), 10))
	}

	createHttpResp, err := r.client.DnsCreateRecordWithResponse(
		ctx,
		data.Domain.ValueString(),
		params,
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create, got error: %s", err))
		return
	} else if createHttpResp.StatusCode() != http.StatusOK || createHttpResp.JSON200 == nil || createHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create, got status code %d: %s", createHttpResp.StatusCode(), string(createHttpResp.Body)))
		return
	}

	id := strconv.FormatInt(createHttpResp.JSON200.Id, 10)
	data.Id = types.StringValue(id)

	readHttpResp, err := r.client.DnsRetrieveRecordsByDomainAndIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		id,
		apiclient.DnsRetrieveRecordsByDomainAndIdJSONRequestBody{
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
	} else if len(readHttpResp.JSON200.Records) != 1 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Expected exactly one record, got %d", len(readHttpResp.JSON200.Records)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, readHttpResp.JSON200.Records[0])...)
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

	httpResp, err := r.client.DnsRetrieveRecordsByDomainAndIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
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
	} else if len(httpResp.JSON200.Records) != 1 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Expected exactly one record, got %d", len(httpResp.JSON200.Records)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, httpResp.JSON200.Records[0])...)
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

	params := apiclient.DnsEditRecordByDomainAndIdJSONRequestBody{
		Apikey:       r.apiKey,
		Secretapikey: r.secretKey,
		Name:         data.Name.ValueStringPointer(),
		Type:         data.Type.ValueString(),
		Content:      data.Content.ValueString(),
	}
	if !data.Ttl.IsNull() {
		params.Ttl = ptr.Ptr(strconv.FormatInt(data.Ttl.ValueInt64(), 10))
	}
	if !data.Priority.IsNull() {
		params.Prio = ptr.Ptr(strconv.FormatInt(data.Priority.ValueInt64(), 10))
	}

	updateHttpResp, err := r.client.DnsEditRecordByDomainAndIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		params,
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update, got error: %s", err))
		return
	} else if updateHttpResp.StatusCode() != http.StatusOK || updateHttpResp.JSON200 == nil || updateHttpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update, got status code %d: %s", updateHttpResp.StatusCode(), string(updateHttpResp.Body)))
		return
	}

	readHttpResp, err := r.client.DnsRetrieveRecordsByDomainAndIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		apiclient.DnsRetrieveRecordsByDomainAndIdJSONRequestBody{
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
	} else if len(readHttpResp.JSON200.Records) != 1 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Expected exactly one record, got %d", len(readHttpResp.JSON200.Records)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, readHttpResp.JSON200.Records[0])...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DnsRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.DnsDeleteRecordByDomainAndIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		apiclient.DomainGetNameServersJSONRequestBody{
			Apikey:       r.apiKey,
			Secretapikey: r.secretKey,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete, got error: %s", err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}
}
