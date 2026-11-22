package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwschema"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

const dnsChangeConfirmationRequiredCode = "DNS_CHANGE_CONFIRMATION_REQUIRED"

type EmailForwardResourceModel struct {
	Domain            types.String `tfsdk:"domain"`
	Address           types.String `tfsdk:"address"`
	ForwardTo         types.String `tfsdk:"forward_to"`
	ConfirmDnsChanges types.Bool   `tfsdk:"confirm_dns_changes"`
}

func NewEmailForwardResource() resource.Resource {
	return &EmailForwardResource{}
}

var _ resource.Resource = &EmailForwardResource{}
var _ resource.ResourceWithConfigure = &EmailForwardResource{}
var _ resource.ResourceWithImportState = &EmailForwardResource{}

type EmailForwardResource struct {
	baseResource
}

func (r *EmailForwardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_forward"
}

func (r *EmailForwardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Describe(ctx, schema.Schema{
		MarkdownDescription: "Manages an email forward for a domain. " +
			"Porkbun identifies a forward by `address` plus `forward_to`; every argument forces replacement.\n\n" +
			"Adding a forward sets Porkbun's MX records and SPF include at the zone apex of the domain's Porkbun DNS. " +
			"If that would delete other apex MX records or rewrite an existing SPF record, creation fails unless `confirm_dns_changes` is `true`. " +
			"Deleting the last forward of a domain also removes Porkbun's forwarding MX and SPF records; this cannot be opted out of.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain to forward email for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				MarkdownDescription: "The address to forward: the part before the `@` (`hello`) or the whole address (`hello@example.com`). Both forms are treated as the same address.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							if req.StateValue.IsNull() || req.PlanValue.IsUnknown() {
								return
							}
							resp.RequiresReplace = normalizeEmailForwardAddress(req.StateValue.ValueString()) !=
								normalizeEmailForwardAddress(req.PlanValue.ValueString())
						},
						"Requires replacement when the local part of the address changes.",
						"Requires replacement when the local part of the address changes.",
					),
				},
			},
			"forward_to": schema.StringAttribute{
				MarkdownDescription: "The mailbox to deliver to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"confirm_dns_changes": schema.BoolAttribute{
				MarkdownDescription: "Agree to the DNS changes Porkbun needs to make at the zone apex, such as deleting other MX records or rewriting an existing SPF record. " +
					"When `false` (the default) and confirmation is required, creation fails and lists the changes. " +
					"This only affects creation and is not read back from the API.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}, fwschema.DefaultSteps()...)
}

func (r *EmailForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EmailForwardResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := apiclient.EmailAddForwardJSONRequestBody{
		Address:   normalizeEmailForwardAddress(data.Address.ValueString()),
		ForwardTo: data.ForwardTo.ValueString(),
	}
	if data.ConfirmDnsChanges.ValueBool() {
		body.ConfirmDnsChanges = new(true)
	}

	httpResp, err := r.client.EmailAddForwardWithResponse(ctx, data.Domain.ValueString(), nil, body)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientCreateError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		if d, ok := dnsChangeConfirmationError(httpResp.Body); ok {
			resp.Diagnostics.Append(d)
			return
		}
		resp.Diagnostics.Append(fwdiag.NewClientCreateHTTPResponseError(httpResp))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// dnsChangeConfirmationError builds a diagnostic listing the DNS changes when
// the API refuses with DNS_CHANGE_CONFIRMATION_REQUIRED.
func dnsChangeConfirmationError(body []byte) (diag.Diagnostic, bool) {
	var e struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		DnsChanges struct {
			MxRecordsRemoved []struct {
				Content string `json:"content"`
				Prio    any    `json:"prio"`
			} `json:"mxRecordsRemoved"`
			Spf       string `json:"spf"`
			SpfBefore string `json:"spfBefore"`
			SpfAfter  string `json:"spfAfter"`
		} `json:"dnsChanges"`
	}
	if err := json.Unmarshal(body, &e); err != nil || e.Code != dnsChangeConfirmationRequiredCode {
		return nil, false
	}

	var b strings.Builder
	b.WriteString("Setting up email forwarding requires changes to the DNS records at the zone apex.")
	if e.Message != "" {
		fmt.Fprintf(&b, " Porkbun said: %s", e.Message)
	}
	b.WriteString("\n\nPlanned changes:\n")
	for _, mx := range e.DnsChanges.MxRecordsRemoved {
		fmt.Fprintf(&b, "  - delete MX record %q (priority %v)\n", mx.Content, mx.Prio)
	}
	if e.DnsChanges.Spf != "" {
		fmt.Fprintf(&b, "  - SPF: %s", e.DnsChanges.Spf)
		if e.DnsChanges.SpfBefore != "" || e.DnsChanges.SpfAfter != "" {
			fmt.Fprintf(&b, " (%q -> %q)", e.DnsChanges.SpfBefore, e.DnsChanges.SpfAfter)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nSet `confirm_dns_changes = true` to accept these changes.")

	return diag.NewErrorDiagnostic("DNS change confirmation required", b.String()), true
}

func (r *EmailForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EmailForwardResourceModel

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

func (r *EmailForwardResource) read(ctx context.Context, data *EmailForwardResourceModel) (found bool, diags diag.Diagnostics) {
	httpResp, err := r.client.GetEmailForwardsWithResponse(ctx, data.Domain.ValueString(), apiclient.GetEmailForwardsJSONRequestBody{})
	if err != nil {
		diags.Append(fwdiag.NewClientReadError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		diags.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	}

	if httpResp.JSON200.Forwards == nil {
		return false, diags
	}
	for _, forward := range *httpResp.JSON200.Forwards {
		if forward.Address == nil || forward.ForwardTo == nil {
			continue
		}
		if !emailForwardMatches(*forward.Address, *forward.ForwardTo, data.Address.ValueString(), data.ForwardTo.ValueString()) {
			continue
		}
		// Keep the configured spelling of address and forward_to when they are
		// equivalent, to avoid spurious diffs; only fill them in on import.
		if data.Address.IsNull() {
			data.Address = types.StringValue(normalizeEmailForwardAddress(*forward.Address))
		}
		if data.ForwardTo.IsNull() {
			data.ForwardTo = types.StringValue(*forward.ForwardTo)
		}
		if data.ConfirmDnsChanges.IsNull() {
			data.ConfirmDnsChanges = types.BoolValue(false)
		}
		return true, diags
	}
	return false, diags
}

// Update only persists in-place changes that do not touch the API: an
// equivalent spelling of address, or confirm_dns_changes.
func (r *EmailForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EmailForwardResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EmailForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EmailForwardResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.EmailDeleteForwardWithResponse(
		ctx,
		data.Domain.ValueString(),
		nil,
		apiclient.EmailDeleteForwardJSONRequestBody{
			Address:   normalizeEmailForwardAddress(data.Address.ValueString()),
			ForwardTo: data.ForwardTo.ValueString(),
		},
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

func (r *EmailForwardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			"Expected import ID in the format '<domain>/<address>/<forward_to>' (e.g. example.com/hello/me@example.org).",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address"), normalizeEmailForwardAddress(parts[1]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("forward_to"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("confirm_dns_changes"), false)...)
}
