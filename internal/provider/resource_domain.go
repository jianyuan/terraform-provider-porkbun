package provider

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
)

type DomainResourceModel struct {
	Domain       types.String `tfsdk:"domain"`
	WhoisPrivacy types.Bool   `tfsdk:"whois_privacy"`
	MaxCost      types.Int64  `tfsdk:"max_cost"`
	Cost         types.Int64  `tfsdk:"cost"`
	OrderId      types.Int64  `tfsdk:"order_id"`
	Status       types.String `tfsdk:"status"`
	Tld          types.String `tfsdk:"tld"`
	CreateDate   types.String `tfsdk:"create_date"`
	ExpireDate   types.String `tfsdk:"expire_date"`
	SecurityLock types.Bool   `tfsdk:"security_lock"`
	AutoRenew    types.Bool   `tfsdk:"auto_renew"`
	ApiAccess    types.Bool   `tfsdk:"api_access"`
	NotLocal     types.Bool   `tfsdk:"not_local"`
}

func (m *DomainResourceModel) Fill(ctx context.Context, domain apiclient.GetDomain200JSONResponseBody_Domain) (diags diag.Diagnostics) {
	m.Domain = types.StringPointerValue(domain.Domain)
	m.Status = types.StringPointerValue(domain.Status)
	m.Tld = types.StringPointerValue(domain.Tld)
	m.CreateDate = types.StringPointerValue(domain.CreateDate)
	m.ExpireDate = types.StringPointerValue(domain.ExpireDate)
	m.SecurityLock = porkbuntypes.FlexibleBoolPointerValue(domain.SecurityLock)
	m.WhoisPrivacy = porkbuntypes.FlexibleBoolPointerValue(domain.WhoisPrivacy)
	m.AutoRenew = porkbuntypes.FlexibleBoolPointerValue(domain.AutoRenew)
	m.ApiAccess = porkbuntypes.FlexibleBoolPointerValue(domain.ApiAccess)
	m.NotLocal = porkbuntypes.FlexibleBoolPointerValue(domain.NotLocal)
	return
}

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

var _ resource.Resource = &DomainResource{}
var _ resource.ResourceWithConfigure = &DomainResource{}
var _ resource.ResourceWithImportState = &DomainResource{}

type DomainResource struct {
	baseResource
}

func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Registers a domain and holds it in state.\n\n" +
			"The price is not written in the configuration. On create the provider quotes the domain " +
			"through `/domain/checkDomain`, converts the quote to the US cents the API demands, and " +
			"sends that as `cost`. Set `max_cost` to put a ceiling on what an unattended apply may spend: " +
			"a first-year promotional price that lapses would otherwise be paid silently at the standard rate.\n\n" +
			"Porkbun exposes no endpoint that deletes a registration, so `terraform destroy` only drops the " +
			"resource from state. The registration itself stands until it expires. Guard the resource with " +
			"`prevent_destroy` if an accidental removal from state would be costly to reconcile.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The fully qualified domain name to register.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"whois_privacy": schema.BoolAttribute{
				MarkdownDescription: "Whether to register with WHOIS privacy. Defaults to the account-level setting. " +
					"A TLD that does not offer privacy ignores this.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"max_cost": schema.Int64Attribute{
				MarkdownDescription: "The most, in US cents, that this registration may cost. The apply fails " +
					"before anything is charged if the live quote is higher. Omit to accept any price.",
				Optional: true,
			},
			"cost": schema.Int64Attribute{
				MarkdownDescription: "What the registration actually cost, in US cents.",
				Computed:            true,
			},
			"order_id": schema.Int64Attribute{
				MarkdownDescription: "Porkbun's internal order id for the registration.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"tld": schema.StringAttribute{
				Computed: true,
			},
			"create_date": schema.StringAttribute{
				Computed: true,
			},
			"expire_date": schema.StringAttribute{
				Computed: true,
			},
			"security_lock": schema.BoolAttribute{
				Computed: true,
			},
			"auto_renew": schema.BoolAttribute{
				Computed: true,
			},
			"api_access": schema.BoolAttribute{
				Computed: true,
			},
			"not_local": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

// quote asks the API what the domain costs right now and returns the total in US
// cents, which is the only form `/domain/create` accepts. The price arrives as a
// decimal string of USD per year, and the registry may demand more than one year,
// so the total is the quoted year multiplied by the minimum duration.
func (r *DomainResource) quote(ctx context.Context, domain string) (cents int64, diags diag.Diagnostics) {
	httpResp, err := r.client.DomainCheckDomainWithResponse(
		ctx,
		domain,
		apiclient.DomainCheckDomainJSONRequestBody{},
	)
	if err != nil {
		diags.Append(fwdiag.NewClientCreateError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
		diags.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	quoted := httpResp.JSON200.Response

	if quoted.Avail == nil || *quoted.Avail != apiclient.CheckDomainResponseResponseAvailYes {
		diags.AddError(
			"Domain not available",
			fmt.Sprintf("%s cannot be registered: the API reports it as unavailable.", domain),
		)
		return
	}

	if quoted.Price == nil {
		diags.AddError("Client error", fmt.Sprintf("The API quoted no registration price for %s.", domain))
		return
	}

	price, err := strconv.ParseFloat(*quoted.Price, 64)
	if err != nil {
		diags.AddError(
			"Client error",
			fmt.Sprintf("The API quoted %q for %s, which is not a price.", *quoted.Price, domain),
		)
		return
	}

	years := int64(1)
	if quoted.MinDuration != nil && *quoted.MinDuration > 1 {
		years = *quoted.MinDuration
	}

	cents = int64(math.Round(price*100)) * years
	return
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := data.Domain.ValueString()

	cost, diags := r.quote(ctx, domain)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refuse before the charge rather than reporting it afterwards. A lapsed
	// first-year promotion is the case this exists for: the apply still
	// succeeds, at ten times the price nobody was watching for.
	if !data.MaxCost.IsNull() && cost > data.MaxCost.ValueInt64() {
		resp.Diagnostics.AddError(
			"Registration costs more than max_cost",
			fmt.Sprintf(
				"Registering %s costs %d US cents, above the %d permitted by max_cost. Nothing was charged.",
				domain, cost, data.MaxCost.ValueInt64(),
			),
		)
		return
	}

	body := apiclient.DomainCreateJSONRequestBody{
		AgreeToTerms: "yes",
		Cost:         cost,
	}
	if !data.WhoisPrivacy.IsNull() && !data.WhoisPrivacy.IsUnknown() {
		body.WhoisPrivacy = data.WhoisPrivacy.ValueBoolPointer()
	}

	httpResp, err := r.client.DomainCreateWithResponse(ctx, domain, body)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	created, err := httpResp.JSON200.AsCreateDomainResponse()
	if err != nil || created.Status != "SUCCESS" {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	data.Cost = types.Int64PointerValue(created.Cost)
	data.OrderId = types.Int64Value(created.OrderId)

	found, diags := r.read(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Client error",
			fmt.Sprintf("%s registered as order %d but does not appear in the account.", domain, created.OrderId),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update carries the plan into state. Everything a registration is made of forces
// replacement, so the only attribute that reaches here is `max_cost`, which
// constrains a future create and never the domain as it stands.
func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Client error",
			fmt.Sprintf("%s is no longer in the account.", data.Domain.ValueString()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainResource) read(ctx context.Context, data *DomainResourceModel) (found bool, diags diag.Diagnostics) {
	httpResp, err := r.client.GetDomainWithResponse(ctx, data.Domain.ValueString(), nil)
	if err != nil {
		diags.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() == http.StatusNotFound {
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status == nil || *httpResp.JSON200.Status != "SUCCESS" {
		diags.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if httpResp.JSON200.Domain == nil {
		return
	}

	diags.Append(data.Fill(ctx, *httpResp.JSON200.Domain)...)
	found = true
	return
}

// Delete drops the resource from state and leaves the registration alone, because
// the API has no endpoint that would end it. Porkbun cancels a registration only
// through a support request inside a short window after purchase, so a destroy
// here is a bookkeeping change and the domain stays ours until it expires.
func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"Domain registration not cancelled",
		fmt.Sprintf(
			"%s was removed from state, but the API cannot cancel a registration. "+
				"The domain stays in the account until it expires on %s.",
			data.Domain.ValueString(), data.ExpireDate.ValueString(),
		),
	)
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}
