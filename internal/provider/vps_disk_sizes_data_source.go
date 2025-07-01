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
	_ datasource.DataSource              = &VPSDiskSizesDataSource{}
	_ datasource.DataSourceWithConfigure = &VPSDiskSizesDataSource{}
)

// NewVPSDiskSizesDataSource is a helper function to simplify the provider implementation.
func NewVPSDiskSizesDataSource() datasource.DataSource {
	return &VPSDiskSizesDataSource{}
}

// VPSDiskSizesDataSource is the data source implementation.
type VPSDiskSizesDataSource struct {
	client *mythicbeasts.Client
}

type VPSDiskSizesDataSourceModel struct {
	HDD []VPSDiskSizeModel `tfsdk:"hdd"`
	SSD []VPSDiskSizeModel `tfsdk:"ssd"`
}

type VPSDiskSizeModel struct {
	Size types.Int64 `tfsdk:"size"`
}

// Metadata returns the data source type name.
func (d *VPSDiskSizesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_disk_sizes"
}

// Schema defines the schema for the data source.
func (d *VPSDiskSizesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hdd": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"size": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
			"ssd": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"size": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *VPSDiskSizesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VPSDiskSizesDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VPSDiskSizesDataSourceModel

	VPSDiskSizes, err := d.client.GetVPSDiskSizes()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Mythic Beasts VPS disk sizes",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, size := range VPSDiskSizes.HDD {
		VPSDiskSizesState := VPSDiskSizeModel{
			Size: types.Int64Value(size),
		}

		state.HDD = append(state.HDD, VPSDiskSizesState)
	}

	for _, size := range VPSDiskSizes.SSD {
		VPSDiskSizesState := VPSDiskSizeModel{
			Size: types.Int64Value(size),
		}

		state.SSD = append(state.SSD, VPSDiskSizesState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *VPSDiskSizesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
