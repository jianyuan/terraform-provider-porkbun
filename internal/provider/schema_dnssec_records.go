package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func dnssecRecordsSchema() superschema.Schema {
	dnssecRecordAttributes := dnssecRecordSchema().Attributes

	return superschema.Schema{
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Retrieves all DNSSEC records associated with the domain at the registry.",
		},
		Attributes: superschema.Attributes{
			"domain": superschema.StringAttribute{
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "The domain name.",
					Required:            true,
				},
			},
			"records": superschema.SuperSetNestedAttributeOf[DnssecRecordsDataSourceModelRecordsItem]{
				DataSource: &schemaD.SetNestedAttribute{
					MarkdownDescription: "All DNSSEC records.",
					Computed:            true,
				},
				Attributes: dnssecRecordAttributes,
			},
		},
	}
}
