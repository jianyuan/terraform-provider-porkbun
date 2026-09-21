package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func urlForwardSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Manages a URL forward for a domain or subdomain.",
		},
		Attributes: superschema.Attributes{
			"id": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The record ID.",
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
			"domain": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The domain for the URL forward.",
					Required:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"subdomain": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The subdomain portion only (optional, leave blank or omit for root domain). Alphanumeric and hyphens only.",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"include_path": superschema.BoolAttribute{
				Resource: &schemaR.BoolAttribute{
					MarkdownDescription: "Whether to append the request URI path to the forwarding destination.",
					Required:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
			"location": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The destination URL to forward to.",
					Required:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"type": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "Redirect kind. `permanent` = HTTP 301; `temporary` = HTTP 302 (default); `masked` = loads the destination in a frame (URL masking). For a precise code — including a 307 temporary redirect — use `redirect_type`.",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("permanent", "temporary", "masked"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"redirect_type": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The exact redirect type; takes precedence over `type` when supplied. `301` = permanent, `302` or `307` = temporary (this is how you request 307), `masked` = URL masking. Omit to derive from type (temporary->302, permanent->301).",
					Optional:            true,
					Computed:            true,
					Validators: []validator.String{
						stringvalidator.OneOf("301", "302", "307", "masked"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"wildcard": superschema.BoolAttribute{
				Resource: &schemaR.BoolAttribute{
					MarkdownDescription: "Whether to also forward all subdomains of the forwarded subdomain.",
					Required:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}
