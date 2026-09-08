package provider

import (
	schemaD "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	schemaR "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	superschema "github.com/orange-cloudavenue/terraform-plugin-framework-superschema"
)

func dnssecRecordSchema() superschema.Schema {
	return superschema.Schema{
		Resource: superschema.SchemaDetails{
			MarkdownDescription: "Manage a DNSSEC DS or key record at the registry. DNSSEC requirements vary by registry — `key_tag`, `alg`, `digest_type`, and `digest` are the minimum required fields. Key data fields are optional and will be omitted if not accepted by the registry.",
		},
		Attributes: superschema.Attributes{
			"domain": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "The domain for the record being created.",
					Required:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"alg": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "DS Data algorithm number (e.g. 13 for ECDSA P-256 SHA-256).",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"digest": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Hex-encoded digest value.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"digest_type": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "Digest type number (e.g. 2 for SHA-256).",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"key_tag": superschema.StringAttribute{
				Common: &schemaR.StringAttribute{
					MarkdownDescription: "DNSSEC key tag.",
				},
				Resource: &schemaR.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				DataSource: &schemaD.StringAttribute{
					Computed: true,
				},
			},
			"key_data_algo": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "Key data algorithm.",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"key_data_flags": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "Key data flags (optional, used when submitting full key data).",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"key_data_protocol": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "Key data protocol.",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"key_data_public_key": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "Key data public key in base64.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"max_sig_life": superschema.StringAttribute{
				Resource: &schemaR.StringAttribute{
					MarkdownDescription: "Maximum signature lifetime in seconds (registry-specific).",
					Optional:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			"pub_key": superschema.StringAttribute{
				DataSource: &schemaD.StringAttribute{
					MarkdownDescription: "Public key.",
					Computed:            true,
				},
			},
		},
	}
}
