// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/paultibbetts/mythicbeasts-client-go"
	mbVPS "github.com/paultibbetts/mythicbeasts-client-go/vps"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &VPSResource{}
	_ resource.ResourceWithConfigure   = &VPSResource{}
	_ resource.ResourceWithImportState = &VPSResource{}
)

// NewVPSResource is a helper function to simplify the provider implementation.
func NewVPSResource() resource.Resource {
	return &VPSResource{}
}

// VPSResource is the resource implementation.
type VPSResource struct {
	client *mythicbeasts.Client
}

// VPSResourceModel maps the resource schema data.
type VPSResourceModel struct {
	Identifier     types.String `tfsdk:"identifier"`
	Product        types.String `tfsdk:"product"`
	Name           types.String `tfsdk:"name"`
	Hostname       types.String `tfsdk:"hostname"`
	SetForwardDNS  types.Bool   `tfsdk:"set_forward_dns"`
	SetReverseDNS  types.Bool   `tfsdk:"set_reverse_dns"`
	UserData       types.String `tfsdk:"user_data"`
	UserDataString types.String `tfsdk:"user_data_string"`
	IPv4Enabled    types.Bool   `tfsdk:"ipv4_enabled"`
	DiskSize       types.Int64  `tfsdk:"disk_size"`
	Image          types.String `tfsdk:"image"`
	SSHKeys        types.String `tfsdk:"ssh_keys"`
	CreateInZone   types.String `tfsdk:"create_in_zone"`

	HostServer types.String  `tfsdk:"host_server"`
	ISOImage   types.String  `tfsdk:"iso_image"`
	Zone       types.Object  `tfsdk:"zone"`
	Family     types.String  `tfsdk:"family"`
	CPUMode    types.String  `tfsdk:"cpu_mode"`
	NetDevice  types.String  `tfsdk:"net_device"`
	DiskBus    types.String  `tfsdk:"disk_bus"`
	Tablet     types.Bool    `tfsdk:"tablet"`
	Price      types.Float64 `tfsdk:"price"`
	Period     types.String  `tfsdk:"period"`
	Dormant    types.Bool    `tfsdk:"dormant"`
	BootDevice types.String  `tfsdk:"boot_device"`
	IPv4       types.Set     `tfsdk:"ipv4"`
	IPv6       types.Set     `tfsdk:"ipv6"`
	Specs      types.Object  `tfsdk:"specs"`
	Macs       types.List    `tfsdk:"macs"`
	SSHProxy   types.Object  `tfsdk:"ssh_proxy"`
	VNC        types.Object  `tfsdk:"vnc"`
}

type ZoneModel struct {
	Code types.String `tfsdk:"code"`
	Name types.String `tfsdk:"name"`
}

type SpecsModel struct {
	DiskType   types.String `tfsdk:"disk_type"`
	DiskSize   types.Int64  `tfsdk:"disk_size"`
	Cores      types.Int64  `tfsdk:"cores"`
	ExtraCores types.Int64  `tfsdk:"extra_cores"`
	RAM        types.Int64  `tfsdk:"ram"`
	ExtraRAM   types.Int64  `tfsdk:"extra_ram"`
}

type VNCModel struct {
	Mode     types.String `tfsdk:"mode"`
	Password types.String `tfsdk:"password"`
	IPv4     types.String `tfsdk:"ipv4"`
	IPv6     types.String `tfsdk:"ipv6"`
	Port     types.Int64  `tfsdk:"port"`
	Display  types.Int64  `tfsdk:"display"`
}

// Metadata returns the resource type name.
func (r *VPSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps"
}

// Schema defines the schema for the resource.
func (r *VPSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Mythic Beasts VPS.\n\n" +
			"In-place updates are supported for `product`, `name`, `disk_size`, `specs.extra_cores`, `specs.extra_ram`, `iso_image`, `boot_device`, `cpu_mode`, `net_device`, `disk_bus`, and `tablet`.\n\n" +
			"The Mythic Beasts API requires the VPS to be powered off before changing `iso_image`, `boot_device`, `cpu_mode`, `net_device`, `disk_bus`, or `tablet`. The provider automatically powers off a running VPS before applying these changes and powers it back on afterwards.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Required: true,
				// can only be between 3 and 20 characters long
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 20),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]*$`),
						"must consist only of lower-case letters and digits",
					),
				},
				MarkdownDescription: "A unique identifier for the server. This will form part of the hostname for the server, and must consist only of lower-case letters and digits and be at most 20 characters long",
			},
			"product": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\S+$`),
						"must not be empty or contain whitespace",
					),
				},
				MarkdownDescription: "Virtual server product code; see the `mythicbeasts_vps_products` data source for valid values",
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\S`),
						"must not be empty or whitespace",
					),
				},
			},
			"hostname": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Hostname the new server should be installed with\nDefault: `{identifier}.vs.mythic-beasts.com`",
			},
			"set_forward_dns": schema.BoolAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Whether to automatically add A/AAAA records for the server's IP addresses to the selected hostname\nDefault: `false`",
			},
			"set_reverse_dns": schema.BoolAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Whether to automatically set reverse DNS for the server's IP addresses to the selected hostname\nDefault: `false`",
			},
			"ipv4_enabled": schema.BoolAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Whether or not to allocate an IPv4 address for this server; an IPv6 address will always be allocated; IPv4 is a chargeable option; see the `mythicbeasts_vps_pricing` data source for the price",
			},
			"disk_size": schema.Int64Attribute{
				Required:            true,
				WriteOnly:           true,
				MarkdownDescription: "Disk size, in MB; see the `mythicbeasts_vps_disk_sizes` data source for valid values",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"image": schema.StringAttribute{
				Required:            true,
				WriteOnly:           true,
				MarkdownDescription: "Operating system image name; see the `mythicbeasts_vps_images` data source for valid values",
			},
			"user_data": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Stored user data ID or name; see the `mythicbeasts_user_data` resource for valid values",
			},
			"user_data_string": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "User data (as a literal string)",
			},
			"ssh_keys": schema.StringAttribute{
				Required:            true,
				WriteOnly:           true,
				MarkdownDescription: "Public SSH key(s) to be added to /root/.ssh/authorized_keys on server",
			},
			"create_in_zone": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Zone (datacentre) code; see the `mythicbeasts_vps_zones` data source for valid values",
			},
			"host_server": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Name of private cloud host server to provision on; see the `mythicbeasts_hosts` data source for valid values",
			},
			"cpu_mode": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString("performance"),
				Validators: []validator.String{
					stringvalidator.OneOf("performance", "compatibility"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "CPU mode.\nPossible values:\n- `performance`\n- `compatibility`\nDefault: `performance`\n\nChanging this setting via the API requires the VPS to be powered off.",
			},
			"net_device": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString("virtio"),
				Validators: []validator.String{
					stringvalidator.OneOf("virtio", "e1000", "rtl8139", "ne2k_pci"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Virtual network device type\nPossible values:\n- `virtio`\n- `e1000`\n-`rtl8139`\n- `ne2k_pci`\nDefault: `virtio`\n\nChanging this setting via the API requires the VPS to be powered off.",
			},
			"disk_bus": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString("virtio"),
				Validators: []validator.String{
					stringvalidator.OneOf("virtio", "sata", "scsi", "ide"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "(Optional) Virtual disk bus adapter type\nPossible values:\n-`virtio`\n-`sata`\n-`scsi`\n-`ide`\nDefault: `virtio`\n\nChanging this setting via the API requires the VPS to be powered off.",
			},
			"tablet": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Tablet mode for VNC mouse pointer\nDefault: `true`\n\nChanging this setting via the API requires the VPS to be powered off.",
			},
			"iso_image": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "ISO image currently in virtual CD drive. Set to `null` to remove.\n\nChanging this setting via the API requires the VPS to be powered off.",
			},
			"family": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Product family code",
			},
			"price": schema.Float64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Price of server (pence per billing period)",
			},
			"period": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Billing period",
			},
			"dormant": schema.BoolAttribute{
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the server is dormant",
			},
			"boot_device": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Boot device.\n\nChanging this setting via the API requires the VPS to be powered off.",
			},
			"ipv4": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "List of IPv4 addresses, if IPv4 was enabled during creation",
			},
			"ipv6": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "List of IPv6 addresses",
			},
			"zone": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"code": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Zone Code",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Zone Name",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Zone (datacentre)",
			},
			"specs": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"disk_type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Disk type",
					},
					"disk_size": schema.Int64Attribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Disk size in MB",
					},
					"cores": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Number of virtual CPU cores",
					},
					"extra_cores": schema.Int64Attribute{
						Computed: true,
						Optional: true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Number of CPU cores in addition to the ones provided by the base product (private cloud only)",
					},
					"ram": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "RAM size in MB",
					},
					"extra_ram": schema.Int64Attribute{
						Computed: true,
						Optional: true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
							MultipleOf(1024),
						},
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Amount of RAM (in MB) in addition to the RAM provided by the base product (private cloud only)",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Server specs",
			},
			"macs": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "List of MAC addresses",
			},
			"ssh_proxy": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "SSH proxy hostname",
					},
					"port": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "SSH proxy port",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "SSH Proxy settings (for IPv4 access to IPv6-only servers)",
			},
			"vnc": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "VNC mode",
					},
					"password": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "VNC password",
					},
					"ipv4": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "VNC IPv4 address",
					},
					"ipv6": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "VNC IPv6 address",
					},
					"port": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "VNC port number",
					},
					"display": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "VNC display number",
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "VNC settings",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *VPSResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mythicbeasts.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mythicbeasts.Client, got: %T. VPS report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *VPSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan VPSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var VPS mbVPS.CreateRequest

	identifier := plan.Identifier.ValueString()

	// get write-only values from the config
	var config VPSResourceModel
	d := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(d...)

	VPS.IPv4 = config.IPv4Enabled.ValueBool()
	VPS.SetForwardDNS = config.SetForwardDNS.ValueBool()
	VPS.SetReverseDNS = config.SetReverseDNS.ValueBool()
	VPS.DiskSize = config.DiskSize.ValueInt64()
	VPS.Image = config.Image.ValueString()

	if !config.SSHKeys.IsNull() && !config.SSHKeys.IsUnknown() {
		VPS.SSHKeys = config.SSHKeys.ValueString()
	}

	if !config.CreateInZone.IsNull() && !config.CreateInZone.IsUnknown() {
		VPS.Zone = config.CreateInZone.ValueString()
	}

	if !config.UserData.IsNull() && !config.UserData.IsUnknown() {
		VPS.UserData = config.UserData.ValueString()
	}

	if !config.UserDataString.IsNull() && !config.UserDataString.IsUnknown() {
		VPS.UserDataString = config.UserDataString.ValueString()
	}

	VPS.Product = plan.Product.ValueString()
	VPS.Name = plan.Name.ValueString()

	if !plan.Hostname.IsNull() && !plan.Hostname.IsUnknown() {
		VPS.Hostname = plan.Hostname.ValueString()
	}

	if !plan.HostServer.IsNull() && !plan.HostServer.IsUnknown() {
		VPS.HostServer = plan.HostServer.ValueString()
	}

	if !plan.CPUMode.IsNull() && !plan.CPUMode.IsUnknown() {
		VPS.CPUMode = plan.CPUMode.ValueString()
	}

	if !plan.NetDevice.IsNull() && !plan.NetDevice.IsUnknown() {
		VPS.NetDevice = plan.NetDevice.ValueString()
	}

	if !plan.DiskBus.IsNull() && !plan.DiskBus.IsUnknown() {
		VPS.DiskBus = plan.DiskBus.ValueString()
	}

	if !plan.Tablet.IsNull() && !plan.Tablet.IsUnknown() {
		VPS.SetTablet(plan.Tablet.ValueBool())
	}

	if extraCores, ok := specInt64FromSpecsObject(plan.Specs, "extra_cores"); ok {
		VPS.ExtraCores = extraCores
	}

	if extraRAM, ok := specInt64FromSpecsObject(plan.Specs, "extra_ram"); ok {
		VPS.ExtraRAM = extraRAM
	}

	VPSJSON, err := json.Marshal(VPS)
	if err != nil {
		tflog.Warn(ctx, "Failed to marshal VPS for logging", map[string]interface{}{"error": err.Error()})
	} else {
		var VPSMap map[string]interface{}
		err = json.Unmarshal(VPSJSON, &VPSMap)
		if err != nil {
			tflog.Warn(ctx, "Failed to unmarshal VPS JSON for logging", map[string]interface{}{"error": err.Error()})
		} else {
			tflog.Info(ctx, "Creating VPS with the following config", VPSMap)
		}
	}

	data, err := r.client.VPS().Create(ctx, identifier, VPS)
	if err != nil {
		var identifierConflictErr *mbVPS.ErrIdentifierConflict
		if errors.As(err, &identifierConflictErr) {
			resp.Diagnostics.AddAttributeError(
				path.Root("identifier"),
				"Identifier already in use",
				fmt.Sprintf("The identifier %q is already in use. Please choose a different one.", plan.Identifier.String()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error creating VPS",
			"Could not create VPS, unexpected error: "+err.Error(),
		)
		return
	}

	server, d := readServer(data)
	resp.Diagnostics.Append(d...)

	diags = resp.State.Set(ctx, *server)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func readServer(server mbVPS.Server) (*VPSResourceModel, diag.Diagnostics) {
	var state VPSResourceModel
	var diags diag.Diagnostics

	state.Identifier = types.StringValue(server.Identifier)
	state.Name = types.StringValue(server.Name)
	state.HostServer = types.StringValue(server.HostServer)
	state.Product = types.StringValue(server.Product)
	state.Family = types.StringValue(server.Family)
	state.CPUMode = types.StringValue(server.CPUMode)
	state.NetDevice = types.StringValue(server.NetDevice)
	state.DiskBus = types.StringValue(server.DiskBus)
	state.Tablet = types.BoolValue(server.Tablet)
	state.Price = types.Float64Value(server.Price)
	state.Period = types.StringValue(server.Period)
	state.Dormant = types.BoolValue(server.Dormant)
	state.BootDevice = types.StringValue(server.BootDevice)
	state.ISOImage = types.StringValue(server.ISOImage)

	ipv4 := []attr.Value{}
	for _, ip := range server.IPv4 {
		ipv4 = append(ipv4, types.StringValue(ip))
	}
	ipv4Val, d := types.SetValue(types.StringType, ipv4)
	diags = append(diags, d...)
	state.IPv4 = ipv4Val

	ipv6 := []attr.Value{}
	for _, ip := range server.IPv6 {
		ipv6 = append(ipv6, types.StringValue(ip))
	}
	ipv6Val, d := types.SetValue(types.StringType, ipv6)
	diags = append(diags, d...)
	state.IPv6 = ipv6Val

	macs := []attr.Value{}
	for _, mac := range server.Macs {
		macs = append(macs, types.StringValue(mac))
	}
	macsVal, d := types.ListValue(types.StringType, macs)
	diags = append(diags, d...)
	state.Macs = macsVal

	zone, d := types.ObjectValue(
		map[string]attr.Type{
			"code": types.StringType,
			"name": types.StringType,
		},
		map[string]attr.Value{
			"code": types.StringValue(server.Zone.Code),
			"name": types.StringValue(server.Zone.Name),
		},
	)
	diags = append(diags, d...)
	state.Zone = zone

	sshProxy, d := types.ObjectValue(
		map[string]attr.Type{
			"hostname": types.StringType,
			"port":     types.Int64Type,
		},
		map[string]attr.Value{
			"hostname": types.StringValue(server.SSHProxy.Hostname),
			"port":     types.Int64Value(server.SSHProxy.Port),
		},
	)
	diags = append(diags, d...)
	state.SSHProxy = sshProxy

	vnc, d := types.ObjectValue(
		map[string]attr.Type{
			"mode":     types.StringType,
			"password": types.StringType,
			"ipv4":     types.StringType,
			"ipv6":     types.StringType,
			"port":     types.Int64Type,
			"display":  types.Int64Type,
		},
		map[string]attr.Value{
			"mode":     types.StringValue(server.VNC.Mode),
			"password": types.StringValue(server.VNC.Password),
			"ipv4":     types.StringValue(server.VNC.IPv4),
			"ipv6":     types.StringValue(server.VNC.IPv6),
			"port":     types.Int64Value(server.VNC.Port),
			"display":  types.Int64Value(server.VNC.Display),
		},
	)
	diags = append(diags, d...)
	state.VNC = vnc

	specs, d := types.ObjectValue(
		map[string]attr.Type{
			"disk_type":   types.StringType,
			"disk_size":   types.Int64Type,
			"cores":       types.Int64Type,
			"extra_cores": types.Int64Type,
			"ram":         types.Int64Type,
			"extra_ram":   types.Int64Type,
		},
		map[string]attr.Value{
			"disk_type":   types.StringValue(server.Specs.DiskType),
			"disk_size":   types.Int64Value(server.Specs.DiskSize),
			"cores":       types.Int64Value(server.Specs.Cores),
			"extra_cores": types.Int64Value(server.Specs.ExtraCores),
			"ram":         types.Int64Value(server.Specs.RAM),
			"extra_ram":   types.Int64Value(server.Specs.ExtraRAM),
		},
	)
	diags = append(diags, d...)
	state.Specs = specs

	return &state, diags
}

// Read refreshes the Terraform state with the latest data.
func (r *VPSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VPSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("reading %s", state.Identifier.ValueString()))

	data, err := r.client.VPS().Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Mythic Beasts VPS",
			"Could not read VPS "+state.Identifier.String()+": "+err.Error(),
		)
		return
	}

	server, d := readServer(data)
	resp.Diagnostics.Append(d...)

	diags = resp.State.Set(ctx, *server)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *VPSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VPSResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config VPSResourceModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VPSResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := mbVPS.NewUpdateRequest()
	hasUpdate := false

	if plan.Product.ValueString() != state.Product.ValueString() {
		updateReq.SetProduct(plan.Product.ValueString())
		hasUpdate = true
	}

	if plan.Name.ValueString() != state.Name.ValueString() {
		updateReq.SetName(plan.Name.ValueString())
		hasUpdate = true
	}

	if !plan.BootDevice.IsNull() && !plan.BootDevice.IsUnknown() && plan.BootDevice.ValueString() != state.BootDevice.ValueString() {
		updateReq.SetBootDevice(plan.BootDevice.ValueString())
		hasUpdate = true
	}

	if !plan.CPUMode.IsNull() && !plan.CPUMode.IsUnknown() && plan.CPUMode.ValueString() != state.CPUMode.ValueString() {
		updateReq.SetCPUMode(plan.CPUMode.ValueString())
		hasUpdate = true
	}

	if !plan.NetDevice.IsNull() && !plan.NetDevice.IsUnknown() && plan.NetDevice.ValueString() != state.NetDevice.ValueString() {
		updateReq.SetNetDevice(plan.NetDevice.ValueString())
		hasUpdate = true
	}

	if !plan.DiskBus.IsNull() && !plan.DiskBus.IsUnknown() && plan.DiskBus.ValueString() != state.DiskBus.ValueString() {
		updateReq.SetDiskBus(plan.DiskBus.ValueString())
		hasUpdate = true
	}

	if !plan.Tablet.IsNull() && !plan.Tablet.IsUnknown() && plan.Tablet.ValueBool() != state.Tablet.ValueBool() {
		updateReq.SetTablet(plan.Tablet.ValueBool())
		hasUpdate = true
	}

	specsUpdate := mbVPS.NewUpdateSpecs()
	hasSpecsUpdate := false

	if !config.DiskSize.IsNull() && !config.DiskSize.IsUnknown() {
		desiredDiskSize := config.DiskSize.ValueInt64()
		currentDiskSize, hasCurrentDiskSize := diskSizeFromSpecsObject(state.Specs)

		if !hasCurrentDiskSize || desiredDiskSize != currentDiskSize {
			specsUpdate.SetDiskSize(desiredDiskSize)
			hasSpecsUpdate = true
		}
	}

	if desiredExtraCores, ok := specInt64FromSpecsObject(plan.Specs, "extra_cores"); ok {
		currentExtraCores, hasCurrentExtraCores := specInt64FromSpecsObject(state.Specs, "extra_cores")
		if !hasCurrentExtraCores || desiredExtraCores != currentExtraCores {
			specsUpdate.SetExtraCores(desiredExtraCores)
			hasSpecsUpdate = true
		}
	}

	if desiredExtraRAM, ok := specInt64FromSpecsObject(plan.Specs, "extra_ram"); ok {
		currentExtraRAM, hasCurrentExtraRAM := specInt64FromSpecsObject(state.Specs, "extra_ram")
		if !hasCurrentExtraRAM || desiredExtraRAM != currentExtraRAM {
			specsUpdate.SetExtraRAM(desiredExtraRAM)
			hasSpecsUpdate = true
		}
	}

	if hasSpecsUpdate {
		updateReq.SetSpecs(specsUpdate)
		hasUpdate = true
	}

	if !plan.ISOImage.IsUnknown() {
		currentISOImage := ""
		if !state.ISOImage.IsNull() && !state.ISOImage.IsUnknown() {
			currentISOImage = state.ISOImage.ValueString()
		}

		if plan.ISOImage.IsNull() {
			if currentISOImage != "" {
				updateReq.ClearISOImage()
				hasUpdate = true
			}
		} else if plan.ISOImage.ValueString() != currentISOImage {
			updateReq.SetISOImage(plan.ISOImage.ValueString())
			hasUpdate = true
		}
	}

	if hasUpdate {
		requiresPoweredOff := updateReq.RequiresPoweredOff()
		shouldPowerOnAfter := false

		if requiresPoweredOff {
			currentServer, err := r.client.VPS().Get(ctx, state.Identifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error reading Mythic Beasts VPS",
					"Could not read VPS "+state.Identifier.String()+": "+err.Error(),
				)
				return
			}

			if strings.EqualFold(currentServer.Status, "running") {
				_, err = r.client.VPS().ShutdownWithGrace(ctx, state.Identifier.ValueString(), 0)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error shutting down VPS",
						"Could not gracefully shut down VPS before update, unexpected error: "+err.Error(),
					)
					return
				}
				shouldPowerOnAfter = true
			}
		}

		_, err := r.client.VPS().Update(ctx, state.Identifier.ValueString(), updateReq)
		if err != nil {
			detail := "Could not update VPS, unexpected error: " + err.Error()
			if shouldPowerOnAfter {
				if _, powerErr := r.client.VPS().SetPower(ctx, state.Identifier.ValueString(), mbVPS.PowerActionOn); powerErr != nil {
					detail += "\n\nThe provider also failed to power the VPS back on: " + powerErr.Error()
				}
			}

			resp.Diagnostics.AddError(
				"Error updating VPS",
				detail,
			)
			return
		}

		if shouldPowerOnAfter {
			_, err = r.client.VPS().SetPower(ctx, state.Identifier.ValueString(), mbVPS.PowerActionOn)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error powering on VPS",
					"The VPS was updated, but powering it back on failed: "+err.Error(),
				)
				return
			}
		}
	}

	updated, err := r.client.VPS().Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Mythic Beasts VPS",
			"Could not read VPS "+state.Identifier.String()+": "+err.Error(),
		)
		return
	}

	server, d := readServer(updated)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, *server)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func diskSizeFromSpecsObject(specs types.Object) (int64, bool) {
	return specInt64FromSpecsObject(specs, "disk_size")
}

func specInt64FromSpecsObject(specs types.Object, key string) (int64, bool) {
	if specs.IsNull() || specs.IsUnknown() {
		return 0, false
	}

	specValue, exists := specs.Attributes()[key]
	if !exists {
		return 0, false
	}

	specAttr, ok := specValue.(types.Int64)
	if !ok || specAttr.IsNull() || specAttr.IsUnknown() {
		return 0, false
	}

	return specAttr.ValueInt64(), true
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *VPSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VPSResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.VPS().Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPS",
			"Could not delete VPS, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *VPSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)

	tflog.Info(ctx, "importing...")
}
