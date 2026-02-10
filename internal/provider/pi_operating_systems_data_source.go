// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/paultibbetts/mythicbeasts-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &piOperatingSystemsDataSource{}
	_ datasource.DataSourceWithConfigure = &piOperatingSystemsDataSource{}
)

// NewPiOperatingSystemsDataSource is a helper function to simplify the provider implementation.
func NewPiOperatingSystemsDataSource() datasource.DataSource {
	return &piOperatingSystemsDataSource{}
}

// piOperatingSystemsDataSource is the data source implementation.
type piOperatingSystemsDataSource struct {
	client *mythicbeasts.Client
}

type piOperatingSystemsDataSourceModel struct {
	Model  types.Int64     `tfsdk:"model"`
	Images []piImagesModel `tfsdk:"images"`
}

type piImagesModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *piOperatingSystemsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pi_operating_systems"
}

// Schema defines the schema for the data source.
func (d *piOperatingSystemsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns operating system images available for a specific Raspberry Pi model. Use this data source to select a valid `os_image` for [`mythicbeasts_pi` resource](../resources/pi).",
		Attributes: map[string]schema.Attribute{
			"model": schema.Int64Attribute{
				Required:    true,
				Description: "Filter for Pi Operating Systems with this model number",
			},
			"images": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *piOperatingSystemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config piOperatingSystemsDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state piOperatingSystemsDataSourceModel

	model := config.Model
	if model.IsNull() || model.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("model"),
			"Missing required model",
			"`model` must be set to read Pi operating systems.",
		)
		return
	}
	state.Model = model

	piOperatingSystems, err := d.client.Pi().GetOperatingSystems(ctx, model.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to read Mythic Beasts Pi operating systems for model: %d", config.Model.ValueInt64()),
			err.Error(),
		)
		return
	}

	// Map response body to model
	for id, name := range piOperatingSystems {
		piOperatingSystemsState := piImagesModel{
			ID:   types.StringValue(id),
			Name: types.StringValue(name),
		}

		state.Images = append(state.Images, piOperatingSystemsState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *piOperatingSystemsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
