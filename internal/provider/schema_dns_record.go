package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func dnsRecordSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Manage a DNS record for a domain.",
		},
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Retrieve a single record for a particular record ID.",
		},
		Attributes: superschema.Attributes{
			"domain": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					Required: true,
				},
				Resource: &schemaR.StringAttribute{
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "The domain name.",
				},
			},
			"id": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The record ID.",
				},
				Resource: &schemaR.StringAttribute{
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Required: true,
				},
			},
			"subdomain": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The subdomain for the record being created, not including the domain itself. Omit to create a record on the root domain. Use * to create a wildcard record.",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"name": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The full name of the record being created, including the subdomain and the domain itself.",
					Computed:            true,
				},
			},
			"type": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The DNS record type.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf(DnsRecordTypes...),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"content": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "The record value.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"ttl": superschema.Int64Attribute{
				Common: &schemaR.Int64Attribute{
					MarkdownDescription: "Time to live in seconds. Minimum is determined by account settings (typically 600). Defaults to the account minimum if omitted or 0.",
					Computed:            true,
				},
				Resource: &schemaR.Int64Attribute{
					Optional: true,
				},
				DataSource: &schemaD.Int64Attribute{},
			},
			"priority": superschema.Int64Attribute{
				Common: &schemaR.Int64Attribute{
					MarkdownDescription: "The priority for `MX` and `SRV` records. Defaults to 0 if omitted..",
					Computed:            true,
				},
				Resource: &schemaR.Int64Attribute{
					Optional: true,
				},
				DataSource: &schemaD.Int64Attribute{},
			},
			"notes": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					Computed: true,
				},
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "Optional notes to store with the record. Not served in DNS.",
					Optional:            true,
				},
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "Notes for the DNS record. Not served in DNS.",
				},
			},
		},
	}
}
