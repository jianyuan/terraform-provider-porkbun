package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwschema"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type WebhookResourceModel struct {
	Id      types.String                  `tfsdk:"id"`
	Url     types.String                  `tfsdk:"url"`
	Events  supertypes.SetValueOf[string] `tfsdk:"events"`
	Enabled types.Bool                    `tfsdk:"enabled"`
	Secret  types.String                  `tfsdk:"secret"`
}

func (m *WebhookResourceModel) FromAPI(ctx context.Context, data apiclient.WebhookEndpoint) (diags diag.Diagnostics) {
	if data.Id != nil {
		m.Id = types.StringValue(strconv.FormatInt(*data.Id, 10))
	}
	m.Url = types.StringPointerValue(data.Url)
	if data.Events != nil {
		diags.Append(m.Events.Set(ctx, *data.Events)...)
	}
	if data.Status != nil {
		m.Enabled = types.BoolValue(*data.Status == "ACTIVE")
	}
	// The secret is only returned to keys with full access; keep the known one otherwise.
	if data.Secret != nil {
		m.Secret = types.StringValue(*data.Secret)
	}
	return
}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

var _ resource.Resource = &WebhookResource{}
var _ resource.ResourceWithConfigure = &WebhookResource{}
var _ resource.ResourceWithImportState = &WebhookResource{}

type WebhookResource struct {
	baseResource
}

func (r *WebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Describe(ctx, schema.Schema{
		MarkdownDescription: "Manages a webhook endpoint that receives signed POST requests for account events. " +
			"Porkbun disables an endpoint after 20 consecutive failed deliveries; the next apply re-enables it while `enabled` is `true`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The webhook endpoint ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The HTTPS URL that receives the signed POST. Port 443 only, and the hostname must resolve to a public internet address.",
				Required:            true,
			},
			"events": schema.SetAttribute{
				MarkdownDescription: "Event types to subscribe to, such as `domain.registered`. Prefix wildcards like `dns.*` are allowed. `*` means all events.",
				CustomType:          supertypes.NewSetTypeOf[string](ctx),
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{types.StringValue("*")})),
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether deliveries are sent.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "The signing secret (HMAC-SHA256 key) used to verify the `X-Porkbun-Signature` header.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}, fwschema.DefaultSteps()...)
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	events := fwdiag.Merge(data.Events.Get(ctx))(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.WebhookCreateWithResponse(ctx, nil, apiclient.WebhookCreateJSONRequestBody{
		Url:    data.Url.ValueString(),
		Events: &events,
	})
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if !apiclient.IsOK(httpResp) || httpResp.JSON200.Endpoint == nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	enabled := data.Enabled
	resp.Diagnostics.Append(data.FromAPI(ctx, *httpResp.JSON200.Endpoint)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Enabled = enabled
	// Save the ID and secret now so a failed disable below does not orphan the endpoint.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create has no status field, so a disabled endpoint needs a follow-up update.
	if !data.Enabled.ValueBool() {
		resp.Diagnostics.Append(r.update(ctx, &data)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, diags := webhookID(data.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.WebhookGetWithResponse(ctx, id, nil)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if !apiclient.IsOK(httpResp) || httpResp.JSON200.Endpoint == nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *httpResp.JSON200.Endpoint)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.update(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) update(ctx context.Context, data *WebhookResourceModel) (diags diag.Diagnostics) {
	id, d := webhookID(data.Id)
	diags.Append(d...)
	events := fwdiag.Merge(data.Events.Get(ctx))(&diags)
	if diags.HasError() {
		return
	}

	status := apiclient.UpdateWebhookRequestStatusDISABLED
	if data.Enabled.ValueBool() {
		status = apiclient.UpdateWebhookRequestStatusACTIVE
	}

	httpResp, err := r.client.WebhookUpdateWithResponse(ctx, nil, apiclient.WebhookUpdateJSONRequestBody{
		Id:     id,
		Url:    data.Url.ValueStringPointer(),
		Events: &events,
		Status: &status,
	})
	if err != nil {
		diags.Append(fwdiag.NewClientUpdateError(err))
		return
	} else if !apiclient.IsOK(httpResp) || httpResp.JSON200.Endpoint == nil {
		diags.Append(fwdiag.NewClientUpdateHTTPResponseError(httpResp))
		return
	}

	diags.Append(data.FromAPI(ctx, *httpResp.JSON200.Endpoint)...)
	return
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, diags := webhookID(data.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.WebhookDeleteWithResponse(ctx, nil, apiclient.WebhookDeleteJSONRequestBody{Id: id})
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

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func webhookID(v types.String) (int64, diag.Diagnostics) {
	id, err := strconv.ParseInt(v.ValueString(), 10, 64)
	if err != nil {
		return 0, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid webhook ID", "Expected a numeric webhook ID, got "+strconv.Quote(v.ValueString())+".")}
	}
	return id, nil
}
