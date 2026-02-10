// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/paultibbetts/mythicbeasts-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &VPSImagesDataSource{}
	_ datasource.DataSourceWithConfigure = &VPSImagesDataSource{}
)

// NewVPSImagesDataSource is a helper function to simplify the provider implementation.
func NewVPSImagesDataSource() datasource.DataSource {
	return &VPSImagesDataSource{}
}

// VPSImagesDataSource is the data source implementation.
type VPSImagesDataSource struct {
	client *mythicbeasts.Client
}

type VPSImagesDataSourceModel struct {
	Images []VPSImageModel `tfsdk:"images"`
}

type VPSImageModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// Metadata returns the data source type name.
func (d *VPSImagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_images"
}

// Schema defines the schema for the data source.
func (d *VPSImagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns available VPS operating system images. Use image names from this data source when creating `mythicbeasts_vps`.",
		Attributes: map[string]schema.Attribute{
			"images": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *VPSImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VPSImagesDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VPSImagesDataSourceModel

	VPSImages, err := d.client.VPS().GetImages(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Mythic Beasts VPS images",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, image := range VPSImages {
		VPSImagesState := VPSImageModel{
			Name:        types.StringValue(image.Name),
			Description: types.StringValue(image.Description),
		}

		state.Images = append(state.Images, VPSImagesState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *VPSImagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mythicbeasts.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mythicbeasts.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}
