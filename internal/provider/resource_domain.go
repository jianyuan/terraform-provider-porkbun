package provider

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwtypes"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

type DomainResourceModel struct {
	MaxCost types.Int64 `tfsdk:"max_cost"`
	Cost    types.Int64 `tfsdk:"cost"`
	OrderId types.Int64 `tfsdk:"order_id"`
	DomainModel
}

func (m *DomainResourceModel) FromAPI(ctx context.Context, domain apiclient.GetDomain200JSONResponseBody_Domain) (diags diag.Diagnostics) {
	diags.Append(m.DomainModel.FromAPI(ctx, domain)...)
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
	resp.Schema = domainSchema().GetResource(ctx)
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

	rat := new(big.Rat)
	_, ok := rat.SetString(*quoted.Price)
	if !ok {
		diags.AddError(
			"Client error",
			fmt.Sprintf("The API quoted %q for %s, which is not a price.", *quoted.Price, domain),
		)
		return
	}

	// Multiply by 100 to convert to cents
	hundred := big.NewRat(100, 1)
	rat.Mul(rat, hundred)

	centsBig := new(big.Int).Div(rat.Num(), rat.Denom())
	if !centsBig.IsInt64() {
		diags.AddError(
			"Client error",
			fmt.Sprintf("The API quoted %q for %s, which is not a price.", *quoted.Price, domain),
		)
		return 0, diags
	}

	// Optionally, multiply by the minimum duration
	years := int64(1)
	if quoted.MinDuration != nil && *quoted.MinDuration > 1 {
		years = *quoted.MinDuration
	}

	centsBig.Mul(centsBig, big.NewInt(years))
	if !centsBig.IsInt64() {
		diags.AddError(
			"Client error",
			fmt.Sprintf("The API quoted %q for %s, which is not a price.", *quoted.Price, domain),
		)
		return 0, diags
	}

	return centsBig.Int64(), diags
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
	if fwtypes.IsKnown(data.MaxCost) && cost > data.MaxCost.ValueInt64() {
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
	if fwtypes.IsKnown(data.WhoisPrivacy) {
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

	diags.Append(data.FromAPI(ctx, *httpResp.JSON200.Domain)...)
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
