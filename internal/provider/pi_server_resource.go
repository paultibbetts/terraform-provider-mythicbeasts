package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	//"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/paultibbetts/mythicbeasts-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &piServerResource{}
	_ resource.ResourceWithConfigure   = &piServerResource{}
	_ resource.ResourceWithImportState = &piServerResource{}
)

// NewPiResource is a helper function to simplify the provider implementation.
func NewPiResource() resource.Resource {
	return &piServerResource{}
}

// piServerResource is the resource implementation.
type piServerResource struct {
	client *mythicbeasts.Client
}

// piServerResourceModel maps the resource schema data.
type piServerResourceModel struct {
	Identifier types.String `tfsdk:"identifier"`
	DiskSize   types.Int64  `tfsdk:"disk_size"`
	SSHKey     types.String `tfsdk:"ssh_key"`
	Model      types.Int64  `tfsdk:"model"`
	Memory     types.Int64  `tfsdk:"memory"`
	CPUSpeed   types.Int64  `tfsdk:"cpu_speed"`
	NICSpeed   types.Int64  `tfsdk:"nic_speed"`
	OSImage    types.String `tfsdk:"os_image"`
	WaitForDNS types.Bool   `tfsdk:"wait_for_dns"`
	IP         types.String `tfsdk:"ip"`
	SSHPort    types.Int64  `tfsdk:"ssh_port"`
	Location   types.String `tfsdk:"location"`
}

// Metadata returns the resource type name.
func (r *piServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pi_server"
}

// Schema defines the schema for the resource.
func (r *piServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Required: true,
				// needs a validator
				// can only be between 3 and 20 characters long
			},
			"disk_size": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				Default:  int64default.StaticInt64(10),
				// needs a validator
				// can only be multiples of 10
			},
			"ssh_key": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"model": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				Default:  int64default.StaticInt64(3),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"memory": schema.Int64Attribute{
				Computed:    false,
				Optional:    true,
				Description: "RAM size in MB. Will default to the lowest available spec matching all of `model`, `memory` and `cpu_speed`.",
				//PlanModifiers: []planmodifier.Int64{
				//	int64planmodifier.UseStateForUnknown(),
				//},
			},
			"cpu_speed": schema.Int64Attribute{
				Computed:    true,
				Optional:    true,
				Description: "CPU speed in MHz. Will default to the lowest available spec matching all of `model`, `memory` and `cpu_speed`.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nic_speed": schema.Int64Attribute{
				Computed:    true,
				Description: "CPU speed in MHz. Only used on creation. Will default to the lowest available spec matching all of `model`, `memory` and `cpu_speed`.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"os_image": schema.StringAttribute{
				Computed: true,
				Optional: true,
				//PlanModifiers: []planmodifier.String{
				//	stringplanmodifier.UseStateForUnknown(),
				//},
			},
			"wait_for_dns": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
			"ip": schema.StringAttribute{
				Computed: true,
				//PlanModifiers: []planmodifier.String{
				//	stringplanmodifier.UseStateForUnknown(),
				//},
			},
			"ssh_port": schema.Int64Attribute{
				Computed: true,
				//PlanModifiers: []planmodifier.Int64{
				//	int64planmodifier.UseStateForUnknown(),
				//},
			},
			"location": schema.StringAttribute{
				Computed: true,
				//PlanModifiers: []planmodifier.String{
				//	stringplanmodifier.UseStateForUnknown(),
				//},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *piServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mythicbeasts.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mythicbeasts.Client, got: %T. piServer report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *piServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan piServerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var piServer mythicbeasts.CreatePiServerRequest

	identifier := plan.Identifier.ValueString()

	if !plan.Model.IsNull() && !plan.Model.IsUnknown() {
		piServer.Model = plan.Model.ValueInt64()
	}

	if !plan.Memory.IsNull() && !plan.Memory.IsUnknown() {
		piServer.Memory = plan.Memory.ValueInt64()
	}

	if !plan.CPUSpeed.IsNull() && !plan.CPUSpeed.IsUnknown() {
		piServer.CPUSpeed = plan.CPUSpeed.ValueInt64()
	}

	if !plan.DiskSize.IsNull() && !plan.DiskSize.IsUnknown() {
		piServer.DiskSize = plan.DiskSize.ValueInt64()
	}

	if !plan.SSHKey.IsNull() && !plan.SSHKey.IsUnknown() {
		piServer.SSHKey = plan.SSHKey.ValueString()
	}

	if !plan.OSImage.IsNull() && !plan.OSImage.IsUnknown() {
		piServer.OSImage = plan.OSImage.ValueString()
	}

	if !plan.WaitForDNS.IsNull() && !plan.WaitForDNS.IsUnknown() {
		piServer.WaitForDNS = plan.WaitForDNS.ValueBool()
	}

	piServerJSON, err := json.Marshal(piServer)
	if err != nil {
		tflog.Warn(ctx, "Failed to marshal PiServer for logging", map[string]interface{}{"error": err.Error()})
	} else {
		var piServerMap map[string]interface{}
		err = json.Unmarshal(piServerJSON, &piServerMap)
		if err != nil {
			tflog.Warn(ctx, "Failed to unmarshal PiServer JSON for logging", map[string]interface{}{"error": err.Error()})
		} else {
			tflog.Info(ctx, "Creating PiServer with the following config", piServerMap)
		}
	}

	// Create new server
	server, err := r.client.CreatePiServer(identifier, piServer)
	if err != nil {
		var identifierConflictErr *mythicbeasts.ErrIdentifierConflict
		if errors.As(err, &identifierConflictErr) {
			resp.Diagnostics.AddAttributeError(
				path.Root("identifier"),
				"Identifier already in use",
				fmt.Sprintf("The identifier %q is already in use. Please choose a different one.", identifier),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error creating Pi server",
			"Could not create Pi server, unexpected error: "+err.Error(),
		)
		return
	}

	// When creating it's `Disk` as an int
	// After creation it's `DiskSize` as a string
	diskSize, err := strconv.ParseFloat(server.DiskSize, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Pi server",
			"Could not create Pi server, unexpected error converting disk size: "+err.Error(),
		)
		return
	}

	state := plan

	tflog.Debug(ctx, "NIC speed in Create", map[string]any{
		"api":   server.NICSpeed,
		"state": state.NICSpeed.ValueInt64(),
	})

	// Map response body to schema and populate Computed attribute values

	state.Memory = types.Int64Value(server.Memory)
	state.CPUSpeed = types.Int64Value(server.CPUSpeed)
	state.NICSpeed = types.Int64Value(server.NICSpeed)
	state.IP = types.StringValue(server.IP)
	state.SSHPort = types.Int64Value(server.SSHPort)
	state.DiskSize = types.Int64Value(int64(diskSize))
	state.Location = types.StringValue(server.Location)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *piServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state piServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetPiServer(state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Mythic Beasts Pi Server",
			"Could not read Pi Server "+state.Identifier.String()+": "+err.Error(),
		)
		return
	}

	diskSize, err := strconv.ParseFloat(server.DiskSize, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Pi server",
			"Could not create Pi server, unexpected error converting disk size: "+err.Error(),
		)
		return
	}

	tflog.Warn(ctx, fmt.Sprintf("memory from api: %d - memory from state: %d", server.Memory, state.Memory.ValueInt64()))
	tflog.Warn(ctx, fmt.Sprintf("ssh port from api: %d - ssh port from state: %d", server.SSHPort, state.SSHPort.ValueInt64()))
	tflog.Warn(ctx, fmt.Sprintf("location from api: %s - location from state: %s", server.Location, state.Location.ValueString()))

	state.Memory = types.Int64Value(server.Memory)
	state.CPUSpeed = types.Int64Value(server.CPUSpeed)
	state.NICSpeed = types.Int64Value(server.NICSpeed)
	state.IP = types.StringValue(server.IP)
	state.SSHPort = types.Int64Value(server.SSHPort)
	state.DiskSize = types.Int64Value(int64(diskSize))
	state.Location = types.StringValue(server.Location)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *piServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *piServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state piServerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePiServer(state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Pi Server",
			"Could not delete Pi Server, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *piServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}
