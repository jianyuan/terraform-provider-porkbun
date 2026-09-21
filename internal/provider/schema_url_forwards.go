package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func urlForwardsSchema() superschema.Schema {
	urlForwardAttributes := urlForwardSchema().Attributes
	delete(urlForwardAttributes, "domain")

	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Retrieve all URL forwards associated with a domain.",
		},
		Attributes: superschema.Attributes{
			"domain": superschema.StringAttribute{
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "The domain name.",
					Required:            true,
				},
			},
			"url_forwards": superschema.SuperSetNestedAttributeOf[URLForwardModel]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "List of URL forwards.",
					Computed:            true,
				},
				Attributes: urlForwardAttributes,
			},
		},
	}
}
