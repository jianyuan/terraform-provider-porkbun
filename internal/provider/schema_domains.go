package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func domainsSchema() superschema.Schema {
	domainAttributes := domainSchema().Attributes
	//nolint:forcetypeassert
	domainAttributes["domain"].(superschema.StringAttribute).DataSource.Required = false
	//nolint:forcetypeassert
	domainAttributes["domain"].(superschema.StringAttribute).DataSource.Computed = true

	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Get all domain names in account.",
		},
		Attributes: superschema.Attributes{
			"domains": superschema.SuperSetNestedAttributeOf[DomainModel]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "Domain names in account.",
					Computed:            true,
				},
				Attributes: domainAttributes,
			},
		},
	}
}
