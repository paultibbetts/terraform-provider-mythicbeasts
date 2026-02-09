// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/paultibbetts/mythicbeasts-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &VPSZonesDataSource{}
	_ datasource.DataSourceWithConfigure = &VPSZonesDataSource{}
)

// NewVPSZonesDataSource is a helper function to simplify the provider implementation.
func NewVPSZonesDataSource() datasource.DataSource {
	return &VPSZonesDataSource{}
}

// VPSZonesDataSource is the data source implementation.
type VPSZonesDataSource struct {
	client *mythicbeasts.Client
}

type VPSZonesDataSourceModel struct {
	Zones []VPSZoneModel `tfsdk:"zones"`
}

type VPSZoneModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Parents     types.List   `tfsdk:"parents"`
}

// Metadata returns the data source type name.
func (d *VPSZonesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_zones"
}

// Schema defines the schema for the data source.
func (d *VPSZonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zones": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"parents": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *VPSZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VPSZonesDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VPSZonesDataSourceModel

	VPSZones, err := d.client.VPS().GetZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Mythic Beasts VPS zones",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, zone := range VPSZones {
		parents := []attr.Value{}
		for _, parent := range zone.Parents {
			parents = append(parents, types.StringValue(parent))
		}
		parentsVal, diags := types.ListValue(types.StringType, parents)
		resp.Diagnostics.Append(diags...)

		VPSZonesState := VPSZoneModel{
			Name:        types.StringValue(zone.Name),
			Description: types.StringValue(zone.Description),
			Parents:     parentsVal,
		}

		state.Zones = append(state.Zones, VPSZonesState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *VPSZonesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
