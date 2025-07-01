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
	_ datasource.DataSource              = &VPSHostsDataSource{}
	_ datasource.DataSourceWithConfigure = &VPSHostsDataSource{}
)

// NewVPSHostsDataSource is a helper function to simplify the provider implementation.
func NewVPSHostsDataSource() datasource.DataSource {
	return &VPSHostsDataSource{}
}

// VPSHostsDataSource is the data source implementation.
type VPSHostsDataSource struct {
	client *mythicbeasts.Client
}

type VPSHostsDataSourceModel struct {
	Hosts []VPSHostInfoModel `tfsdk:"hosts"`
}

type VPSHostInfoModel struct {
	Name     types.String         `tfsdk:"name"`
	Cores    types.Int64          `tfsdk:"cores"`
	RAM      types.Int64          `tfsdk:"ram"`
	Disk     VPSHostDiskInfoModel `tfsdk:"disk"`
	FreeRAM  types.Int64          `tfsdk:"free_ram"`
	FreeDisk VPSHostDiskInfoModel `tfsdk:"free_disk"`
}

type VPSHostDiskInfoModel struct {
	SSD types.Int64 `tfsdk:"ssd"`
	HDD types.Int64 `tfsdk:"hdd"`
}

// Metadata returns the data source type name.
func (d *VPSHostsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_hosts"
}

// Schema defines the schema for the data source.
func (d *VPSHostsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"hosts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Computed: true},
						"cores":    schema.StringAttribute{Computed: true},
						"ram":      schema.Int64Attribute{Computed: true},
						"free_ram": schema.Int64Attribute{Computed: true},
						"disk": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"ssd": schema.Int64Attribute{Computed: true},
								"hdd": schema.Int64Attribute{Computed: true},
							},
						},
						"free_disk": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"ssd": schema.Int64Attribute{Computed: true},
								"hdd": schema.Int64Attribute{Computed: true},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *VPSHostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VPSHostsDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	VPSHosts, err := d.client.GetVPSHosts()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Mythic Beasts VPS zones",
			err.Error(),
		)
		return
	}

	var state VPSHostsDataSourceModel

	for _, host := range *VPSHosts {
		disk := VPSHostDiskInfoModel{
			SSD: types.Int64Value(int64(host.Disk.SSD)),
			HDD: types.Int64Value(host.Disk.HDD),
		}
		freeDisk := VPSHostDiskInfoModel{
			SSD: types.Int64Value(host.FreeDisk.SSD),
			HDD: types.Int64Value(host.FreeDisk.HDD),
		}
		VPSHostState := VPSHostInfoModel{
			Name:     types.StringValue(host.Name),
			Cores:    types.Int64Value(host.Cores),
			RAM:      types.Int64Value(host.RAM),
			FreeRAM:  types.Int64Value(host.FreeRAM),
			Disk:     disk,
			FreeDisk: freeDisk,
		}

		state.Hosts = append(state.Hosts, VPSHostState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *VPSHostsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
