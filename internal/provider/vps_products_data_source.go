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
	_ datasource.DataSource              = &VPSProductsDataSource{}
	_ datasource.DataSourceWithConfigure = &VPSProductsDataSource{}
)

// NewVPSProductsDataSource is a helper function to simplify the provider implementation.
func NewVPSProductsDataSource() datasource.DataSource {
	return &VPSProductsDataSource{}
}

// VPSProductsDataSource is the data source implementation.
type VPSProductsDataSource struct {
	client *mythicbeasts.Client
}

type VPSProductsDataSourceModel struct {
	Products []VPSProductModel `tfsdk:"products"`
}

type VPSProductModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Code        types.String `tfsdk:"code"`
	Family      types.String `tfsdk:"family"`
	Period      types.String `tfsdk:"period"`
	Specs       VPSProductSpecsModel `tfsdk:"specs"`
}

type VPSProductSpecsModel struct {
	Cores     types.Int64 `tfsdk:"cores"`
	RAM       types.Int64 `tfsdk:"ram"`
	Bandwidth types.Int64 `tfsdk:"bandwidth"`
}

// Metadata returns the data source type name.
func (d *VPSProductsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_products"
}

// Schema defines the schema for the data source.
func (d *VPSProductsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"products": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"code": schema.StringAttribute{
							Computed: true,
						},
						"family": schema.StringAttribute{
							Computed: true,
						},
						"period": schema.StringAttribute{
							Computed: true,
						},
						"specs": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"cores": schema.Int64Attribute{
									Computed: true,
								},
								"ram": schema.Int64Attribute{
									Computed: true,
								},
								"bandwidth": schema.Int64Attribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *VPSProductsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VPSProductsDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VPSProductsDataSourceModel

	VPSProducts, err := d.client.VPS().GetProducts(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Mythic Beasts VPS products",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, product := range VPSProducts {
		VPSProductState := VPSProductModel{
			Name:        types.StringValue(product.Name),
			Description: types.StringValue(product.Description),
			Code:        types.StringValue(product.Code),
			Family:      types.StringValue(product.Family),
			Period:      types.StringValue(product.Period),
			Specs: VPSProductSpecsModel{
				Cores:     types.Int64Value(int64(product.Specs.Cores)),
				RAM:       types.Int64Value(int64(product.Specs.RAM)),
				Bandwidth: types.Int64Value(int64(product.Specs.Bandwidth)),
			},
		}

		state.Products = append(state.Products, VPSProductState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *VPSProductsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
