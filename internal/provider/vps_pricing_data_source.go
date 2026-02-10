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
	_ datasource.DataSource              = &VPSPricingDataSource{}
	_ datasource.DataSourceWithConfigure = &VPSPricingDataSource{}
)

// NewVPSPricingDataSource is a helper function to simplify the provider implementation.
func NewVPSPricingDataSource() datasource.DataSource {
	return &VPSPricingDataSource{}
}

// VPSPricingDataSource is the data source implementation.
type VPSPricingDataSource struct {
	client *mythicbeasts.Client
}

type VPSPricingDataSourceModel struct {
	Disk     *VPSDiskPricesModel    `tfsdk:"disk"`
	IPv4     types.Int64            `tfsdk:"ipv4"`
	Products map[string]types.Int64 `tfsdk:"products"`
}
type VPSDiskPricesModel struct {
	SSD VPSDiskPricingModel `tfsdk:"ssd"`
	HDD VPSDiskPricingModel `tfsdk:"hdd"`
}

type VPSDiskPricingModel struct {
	Price  types.Int64 `tfsdk:"price"`
	Extent types.Int64 `tfsdk:"extent"`
}

// Metadata returns the data source type name.
func (d *VPSPricingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_pricing"
}

// Schema defines the schema for the data source.
func (d *VPSPricingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns VPS pricing for disk, IPv4, and product codes. Use this to estimate monthly costs before creating or updating `mythicbeasts_vps`.",
		Attributes: map[string]schema.Attribute{
			"disk": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ssd": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"price": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "Price (in pence for month) per unit of SSD space",
							},
							"extent": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "GB per unit of SSD space",
							},
						},
					},
					"hdd": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"price": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "Price (in pence for month) per unit of HDD space",
							},
							"extent": schema.Int64Attribute{
								Computed:            true,
								MarkdownDescription: "GB per unit of HDD space",
							},
						},
					},
				},
			},
			"ipv4": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Price (in pence per month) for one IPv4 address",
			},
			"products": schema.MapAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *VPSPricingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VPSPricingDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	VPSPricing, err := d.client.VPS().GetPricing(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Mythic Beasts VPS pricing",
			err.Error(),
		)
		return
	}

	products := make(map[string]types.Int64, len(VPSPricing.Products))
	for name, price := range VPSPricing.Products {
		products[name] = types.Int64Value(price)
	}

	SSD := VPSDiskPricingModel{
		Price:  types.Int64Value(VPSPricing.Disk.SSD.Price),
		Extent: types.Int64Value(VPSPricing.Disk.SSD.Extent),
	}

	HDD := VPSDiskPricingModel{
		Price:  types.Int64Value(VPSPricing.Disk.HDD.Price),
		Extent: types.Int64Value(VPSPricing.Disk.HDD.Extent),
	}

	state := VPSPricingDataSourceModel{
		Disk: &VPSDiskPricesModel{
			SSD: SSD,
			HDD: HDD,
		},
		IPv4:     types.Int64Value(VPSPricing.IPv4),
		Products: products,
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *VPSPricingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
