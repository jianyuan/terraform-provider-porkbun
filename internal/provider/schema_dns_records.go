package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func dnsRecordsSchema() superschema.Schema {
	dnsRecordAttributes := dnsRecordSchema().Attributes
	delete(dnsRecordAttributes, "domain")
	delete(dnsRecordAttributes, "subdomain")
	//nolint:forcetypeassert
	dnsRecordAttributes["id"].(superschema.StringAttribute).DataSource.Required = false
	//nolint:forcetypeassert
	dnsRecordAttributes["id"].(superschema.StringAttribute).DataSource.Computed = true

	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Retrieve all editable DNS records associated with a domain.",
		},
		Attributes: superschema.Attributes{
			"domain": superschema.StringAttribute{
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "The domain name.",
					Required:            true,
				},
			},
			"filter": superschema.SuperSingleNestedAttributeOf[DnsRecordsFilterDataSourceModel]{
				DataSource: &schemaD.SingleNestedAttribute{
					MarkdownDescription: "Record filter.",
					Optional:            true,
				},
				Attributes: superschema.Attributes{
					"type": superschema.StringAttribute{
						DataSource: &schemaD.StringAttribute{
							MarkdownDescription: "Record type.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(DnsRecordTypes...),
							},
						},
					},
					"subdomain": superschema.StringAttribute{
						DataSource: &schemaD.StringAttribute{
							MarkdownDescription: "Record subdomain.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("type")),
							},
						},
					},
				},
			},
			"records": superschema.SuperSetNestedAttributeOf[DnsRecordModel]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "All editable DNS records.",
					Computed:            true,
				},
				Attributes: dnsRecordAttributes,
			},
		},
	}
}
