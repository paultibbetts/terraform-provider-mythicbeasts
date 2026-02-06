// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/paultibbetts/mythicbeasts-client-go"
	mbPi "github.com/paultibbetts/mythicbeasts-client-go/pi"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &PiResource{}
	_ resource.ResourceWithConfigure   = &PiResource{}
	_ resource.ResourceWithImportState = &PiResource{}
)

// NewPiResource is a helper function to simplify the provider implementation.
func NewPiResource() resource.Resource {
	return &PiResource{}
}

// PiResource is the resource implementation.
type PiResource struct {
	client *mythicbeasts.Client
}

// PiResourceModel maps the resource schema data.
type PiResourceModel struct {
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
func (r *PiResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pi"
}

// Schema defines the schema for the resource.
func (r *PiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 20),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9\-]*$`),
						"must consist only of alphanumerics and -",
					),
				},
				MarkdownDescription: "A unique identifier for the server. This will form part of the hostname for the server, and should consist only of alphanumerics and `-`.",
			},
			"disk_size": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(10),
				MarkdownDescription: "Disk space size, in GB. Must be a multiple of 10",
				Validators: []validator.Int64{
					MultipleOfTen(),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ssh_key": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Public SSH key(s) to be added to /root/.ssh/authorized_keys on server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				Default:  int64default.StaticInt64(3),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(3, 4),
				},
				MarkdownDescription: "Raspberry Pi model (3 or 4)",
			},
			"memory": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				MarkdownDescription: "RAM size in MB. Will default to the lowest available spec matching all of `model`, `memory` and `cpu_speed`.",
			},
			"cpu_speed": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "CPU speed in MHz. Will default to the lowest available spec matching all of `model`, `memory` and `cpu_speed`.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nic_speed": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "CPU speed in MHz. Only used on creation. Will default to the lowest available spec matching all of `model`, `memory` and `cpu_speed`.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"os_image": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Operating system image",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wait_for_dns": schema.BoolAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Whether to wait for DNS records under hostedpi.com to become available before completing provisioning.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IPv6 address for server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssh_port": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Port for accessing SSH via IPv4 relay. Server is accessible on `ssh.{identifier}.hostedpi.com`.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"location": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data centre in which server is located",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *PiResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mythicbeasts.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mythicbeasts.Client, got: %T. Pi report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *PiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan PiResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config PiResourceModel
	d := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	var Pi mbPi.CreateRequest

	identifier := plan.Identifier.ValueString()

	if !config.SSHKey.IsNull() && !config.SSHKey.IsUnknown() {
		Pi.SSHKey = config.SSHKey.ValueString()
	}

	if !plan.Model.IsNull() && !plan.Model.IsUnknown() {
		Pi.Model = plan.Model.ValueInt64()
	}

	if !plan.Memory.IsNull() && !plan.Memory.IsUnknown() {
		Pi.Memory = plan.Memory.ValueInt64()
	}

	if !plan.CPUSpeed.IsNull() && !plan.CPUSpeed.IsUnknown() {
		Pi.CPUSpeed = plan.CPUSpeed.ValueInt64()
	}

	if !plan.DiskSize.IsNull() && !plan.DiskSize.IsUnknown() {
		Pi.DiskSize = plan.DiskSize.ValueInt64()
	}

	if !plan.OSImage.IsNull() && !plan.OSImage.IsUnknown() {
		Pi.OSImage = plan.OSImage.ValueString()
	}

	waitForDNS := false
	if !config.WaitForDNS.IsNull() && !config.WaitForDNS.IsUnknown() {
		waitForDNS = config.WaitForDNS.ValueBool()
	}
	Pi.WaitForDNS = waitForDNS

	PiJSON, err := json.Marshal(Pi)
	if err != nil {
		tflog.Warn(ctx, "Failed to marshal Pi for logging", map[string]interface{}{"error": err.Error()})
	} else {
		var PiMap map[string]interface{}
		err = json.Unmarshal(PiJSON, &PiMap)
		if err != nil {
			tflog.Warn(ctx, "Failed to unmarshal Pi JSON for logging", map[string]interface{}{"error": err.Error()})
		} else {
			tflog.Info(ctx, "Creating Pi with the following config", PiMap)
		}
	}

	// Create new server
	server, err := r.client.Pi().Create(ctx, identifier, Pi)
	if err != nil {
		var identifierConflictErr *mbPi.ErrIdentifierConflict
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

	// Map response body to schema and populate Computed attribute values

	state.Memory = types.Int64Value(server.Memory)
	state.CPUSpeed = types.Int64Value(server.CPUSpeed)
	state.NICSpeed = types.Int64Value(server.NICSpeed)
	ip, err := normalizeIPv6(server.IP)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Pi server",
			fmt.Sprintf("Could not create Pi server, invalid IPv6 address %q: %s", server.IP, err.Error()),
		)
		return
	}
	state.IP = types.StringValue(ip)
	state.SSHPort = types.Int64Value(server.SSHPort)
	state.DiskSize = types.Int64Value(int64(diskSize))
	state.Location = types.StringValue(server.Location)
	state.Model = types.Int64Value(server.Model)
	state.WaitForDNS = types.BoolValue(waitForDNS)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *PiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PiResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.Pi().Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Mythic Beasts Pi ",
			"Could not read Pi  "+state.Identifier.String()+": "+err.Error(),
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

	tflog.Warn(ctx, fmt.Sprintf("memory from aPi: %d - memory from state: %d", server.Memory, state.Memory.ValueInt64()))
	tflog.Warn(ctx, fmt.Sprintf("ssh port from aPi: %d - ssh port from state: %d", server.SSHPort, state.SSHPort.ValueInt64()))
	tflog.Warn(ctx, fmt.Sprintf("location from aPi: %s - location from state: %s", server.Location, state.Location.ValueString()))

	state.Memory = types.Int64Value(server.Memory)
	state.CPUSpeed = types.Int64Value(server.CPUSpeed)
	state.NICSpeed = types.Int64Value(server.NICSpeed)
	ip, err := normalizeIPv6(server.IP)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Mythic Beasts Pi",
			fmt.Sprintf("Could not read Pi %s, invalid IPv6 address %q: %s", state.Identifier.String(), server.IP, err.Error()),
		)
		return
	}
	state.IP = types.StringValue(ip)
	state.SSHPort = types.Int64Value(server.SSHPort)
	state.DiskSize = types.Int64Value(int64(diskSize))
	state.Location = types.StringValue(server.Location)
	state.Model = types.Int64Value(server.Model)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *PiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *PiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PiResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Pi().Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Pi ",
			"Could not delete Pi , unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *PiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}
