// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/paultibbetts/mythicbeasts-client-go"
	mbVPS "github.com/paultibbetts/mythicbeasts-client-go/vps"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &UserDataResource{}
	_ resource.ResourceWithConfigure   = &UserDataResource{}
	_ resource.ResourceWithImportState = &UserDataResource{}
)

// NewUserDataResource is a helper function to simplify the provider implementation.
func NewUserDataResource() resource.Resource {
	return &UserDataResource{}
}

// UserDataResource is the resource implementation.
type UserDataResource struct {
	client *mythicbeasts.Client
}

// UserDataResourceModel maps the resource schema data.
type UserDataResourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Size types.Int64  `tfsdk:"size"`
	Data types.String `tfsdk:"data"`
}

// Metadata returns the resource type name.
func (r *UserDataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_data"
}

// Schema defines the schema for the resource.
func (r *UserDataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "User data identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "User data name",
			},
			"data": schema.StringAttribute{
				Required: true,
				// needs a validator
				// less than 64kb
				MarkdownDescription: "User data snippet",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "User data size (in bytes)",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *UserDataResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mythicbeasts.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mythicbeasts.Client, got: %T. UserData report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *UserDataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan UserDataResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Data.IsNull() || plan.Data.IsUnknown() {
		resp.Diagnostics.AddError(
			"You must set the data",
			"Data must be set to create a User Data resource",
		)
		return
	}

	var UserData mbVPS.NewUserData

	UserData.Name = plan.Name.ValueString()
	UserData.Data = plan.Data.ValueString()

	UserDataJSON, err := json.Marshal(UserData)
	if err != nil {
		tflog.Warn(ctx, "Failed to marshal User Data for logging", map[string]interface{}{"error": err.Error()})
	} else {
		var UserDataMap map[string]any
		err = json.Unmarshal(UserDataJSON, &UserDataMap)
		if err != nil {
			tflog.Warn(ctx, "Failed to unmarshal UserData JSON for logging", map[string]any{"error": err.Error()})
		}
	}

	_, err = r.client.VPS().CreateUserData(ctx, UserData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating UserData",
			"Could not create UserData, unexpected error: "+err.Error(),
		)
		return
	}

	created, err := r.client.VPS().GetUserDataByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating UserData",
			"Could not create UserData, could not fetch the created resource:"+err.Error(),
		)
		return
	}

	var state UserDataResourceModel

	state.Name = types.StringValue(created.Name)
	state.Data = types.StringValue(created.Data)
	state.ID = types.Int64Value(created.ID)
	state.Size = types.Int64Value(created.Size)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *UserDataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserDataResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.VPS().GetUserData(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Mythic Beasts User Data",
			fmt.Sprintf("Could not read User Data %d: %s", state.ID.ValueInt64(), err.Error()),
		)
		return
	}

	state.ID = types.Int64Value(data.ID)
	state.Name = types.StringValue(data.Name)
	state.Data = types.StringValue(data.Data)
	state.Size = types.Int64Value(data.Size)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *UserDataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *UserDataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserDataResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.VPS().DeleteUserData(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting User Data",
			"Could not delete User Data, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *UserDataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Could not parse ID as int: "+err.Error())
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(id))
}
