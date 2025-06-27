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
	_ datasource.DataSource              = &piModelsDataSource{}
	_ datasource.DataSourceWithConfigure = &piModelsDataSource{}
)

// NewPiModelsDataSource is a helper function to simplify the provider implementation.
func NewPiModelsDataSource() datasource.DataSource {
	return &piModelsDataSource{}
}

// piModelsDataSource is the data source implementation.
type piModelsDataSource struct {
	client *mythicbeasts.Client
}

type piModelsDataSourceModel struct {
	Models   []piModelsModel `tfsdk:"models"`
	Model    types.Int64     `tfsdk:"model"`
	Memory   types.Int64     `tfsdk:"memory"`
	NICSpeed types.Int64     `tfsdk:"nic_speed"`
	CPUSpeed types.Int64     `tfsdk:"cpu_speed"`
}

type piModelsModel struct {
	Model    types.Int64 `tfsdk:"model"`
	Memory   types.Int64 `tfsdk:"memory"`
	NICSpeed types.Int64 `tfsdk:"nic_speed"`
	CPUSpeed types.Int64 `tfsdk:"cpu_speed"`
}

// Metadata returns the data source type name.
func (d *piModelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pi_models"
}

// Schema defines the schema for the data source.
func (d *piModelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"model": schema.Int64Attribute{
				Optional:    true,
				Description: "Filter for Pi models with this model number",
			},
			"memory": schema.Int64Attribute{
				Optional:    true,
				Description: "Filter for Pi models with this much memory (in MB)",
			},
			"nic_speed": schema.Int64Attribute{
				Optional:    true,
				Description: "Filter for Pi models with this NIC speed (in Mbps)",
			},
			"cpu_speed": schema.Int64Attribute{
				Optional:    true,
				Description: "Filter for Pi models with this CPU speed (in MHz)",
			},
			"models": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"model": schema.Int64Attribute{
							Computed: true,
						},
						"memory": schema.Int64Attribute{
							Computed: true,
						},
						"nic_speed": schema.Int64Attribute{
							Computed: true,
						},
						"cpu_speed": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *piModelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config piModelsDataSourceModel // for input
	configDiags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(configDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state piModelsDataSourceModel

	piModels, err := d.client.GetPiModels()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Mythic Beasts Pi models",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, model := range *piModels {
		if !config.Model.IsNull() && model.Model != config.Model.ValueInt64() {
			continue
		}

		if !config.Memory.IsNull() && model.Memory != config.Memory.ValueInt64() {
			continue
		}

		if !config.NICSpeed.IsNull() && model.NICSpeed != config.NICSpeed.ValueInt64() {
			continue
		}

		if !config.CPUSpeed.IsNull() && model.CPUSpeed != config.CPUSpeed.ValueInt64() {
			continue
		}

		piModelsState := piModelsModel{
			Model:    types.Int64Value(model.Model),
			Memory:   types.Int64Value(model.Memory),
			NICSpeed: types.Int64Value(model.NICSpeed),
			CPUSpeed: types.Int64Value(model.CPUSpeed),
		}

		state.Models = append(state.Models, piModelsState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *piModelsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
