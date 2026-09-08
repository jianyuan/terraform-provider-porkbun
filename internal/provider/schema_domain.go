package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func domainSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Registers a domain and holds it in state.\n\n" +
				"The price is not written in the configuration. On create the provider quotes the domain " +
				"through `/domain/checkDomain`, converts the quote to the US cents the API demands, and " +
				"sends that as `cost`. Set `max_cost` to put a ceiling on what an unattended apply may spend: " +
				"a first-year promotional price that lapses would otherwise be paid silently at the standard rate.\n\n" +
				"Porkbun exposes no endpoint that deletes a registration, so `terraform destroy` only drops the " +
				"resource from state. The registration itself stands until it expires. Guard the resource with " +
				"`prevent_destroy` if an accidental removal from state would be costly to reconcile.",
		},
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Get the metadata for a single domain.",
		},
		Attributes: superschema.Attributes{
			"domain": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The fully qualified domain name to register.",
					Required:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "Fully qualified domain name in the authenticated account.",
					Required:            true,
				},
			},
			"max_cost": superschema.Int64Attribute{
				Resource: &schemaR.Int64Attribute{
					MarkdownDescription: "The most, in US cents, that this registration may cost. The apply fails " +
						"before anything is charged if the live quote is higher. Omit to accept any price.",
					Optional: true,
				},
			},
			"cost": superschema.Int64Attribute{
				Resource: &schemaR.Int64Attribute{
					MarkdownDescription: "What the registration actually cost, in US cents.",
					Computed:            true,
				},
			},
			"order_id": superschema.Int64Attribute{
				Resource: &schemaR.Int64Attribute{
					MarkdownDescription: "Porkbun's internal order id for the registration.",
					Computed:            true,
				},
			},
			"status": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The status of the domain.",
					Computed:            true,
				},
			},
			"tld": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The TLD of the domain.",
					Computed:            true,
				},
			},
			"create_date": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The creation date of the domain.",
					Computed:            true,
				},
			},
			"expire_date": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The expiration date of the domain.",
					Computed:            true,
				},
			},
			"security_lock": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "Indicates if the Registrar Lock/Transfer Lock is enabled to prevent unauthorized domain transfers.",
					Computed:            true,
				},
			},
			"whois_privacy": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "Indicates if WHOIS privacy protection is active to hide personal registrant details from the public WHOIS database.",
					Computed:            true,
				},
				Resource: &schemaR.BoolAttribute{
					MarkdownDescription: "Whether to register with WHOIS privacy. Defaults to the account-level setting. A TLD that does not offer privacy ignores this.",
					Optional:            true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
			"auto_renew": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "Indicates whether the domain will automatically renew before its expiration date.",
					Computed:            true,
				},
			},
			"api_access": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "Indicates whether API access is enabled for the domain.",
					Computed:            true,
				},
			},
			"not_local": superschema.BoolAttribute{
				Common: &schemaR.BoolAttribute{
					MarkdownDescription: "Indicates if the domain is registered with an external, third-party registrar.",
					Computed:            true,
				},
			},
		},
	}
}
