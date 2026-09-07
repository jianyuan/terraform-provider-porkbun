package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func domainNameserversSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Update the name servers for your domain.",
		},
		DataSource: superschema.SchemaDetails{
			MarkdownDescription: "Get the authoritative name servers listed at the registry for your domain.",
		},
		Attributes: superschema.Attributes{
			"domain": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					Required: true,
				},
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The domain name for which to manage the authoritative name servers.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "The domain name for which to retrieve the authoritative name servers.",
				},
			},
			"nameservers": superschema.SuperSetAttributeOf[string]{
				Common: &schemaR.SetAttribute{
					MarkdownDescription: "The authoritative name servers for the domain.",
				},
				Resource: &schemaR.SetAttribute{
					Required: true,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
				},
				DataSource: &schemaD.SetAttribute{
					Computed: true,
				},
			},
		},
	}
}
