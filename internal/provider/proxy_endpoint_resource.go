// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/paultibbetts/mythicbeasts-client-go"
	mbProxy "github.com/paultibbetts/mythicbeasts-client-go/proxy"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ProxyEndpointResource{}
	_ resource.ResourceWithConfigure   = &ProxyEndpointResource{}
	_ resource.ResourceWithImportState = &ProxyEndpointResource{}
)

// NewProxyEndpointResource is a helper function to simplify the provider implementation.
func NewProxyEndpointResource() resource.Resource {
	return &ProxyEndpointResource{}
}

// ProxyEndpointResource is the resource implementation.
type ProxyEndpointResource struct {
	client *mythicbeasts.Client
}

// ProxyEndpointResourceModel maps the resource schema data.
type ProxyEndpointResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Domain        types.String `tfsdk:"domain"`
	Hostname      types.String `tfsdk:"hostname"`
	Address       types.String `tfsdk:"address"`
	Site          types.String `tfsdk:"site"`
	ProxyProtocol types.Bool   `tfsdk:"proxy_protocol"`
}

// Metadata returns the resource type name.
func (r *ProxyEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_endpoint"
}

// Schema defines the schema for the resource.
func (r *ProxyEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Endpoints for the IPv4 to IPv6 proxy.\n\n" +
			"Can be used to make [`mythicbeasts_pi` resources](../resources/pi) available via IPv4.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Composite identifier for the proxy endpoint (domain/hostname/address/site).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Domain part of the hostname to be proxied (e.g. \"example.com\").",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\S+$`),
						"must not be empty or contain whitespace",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Host part of the hostname to be proxied (e.g. \"www\" or \"@\").",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\S+$`),
						"must not be empty or contain whitespace",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IPv6 address of the server to which requests are proxied.",
				PlanModifiers: []planmodifier.String{
					IPv6Normalize(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Site in which the proxy server is located, or `all` for all sites.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\S+$`),
						"must not be empty or contain whitespace",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy_protocol": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether PROXY protocol is enabled for this endpoint.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProxyEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mythicbeasts.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mythicbeasts.Client, got: %T. Proxy Endpoint report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *ProxyEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan ProxyEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config ProxyEndpointResourceModel
	d := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()
	hostname := plan.Hostname.ValueString()
	address := plan.Address.ValueString()
	normalizedAddress, err := normalizeIPv6(address)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("address"), "Invalid IPv6 address", err.Error())
		return
	}
	address = normalizedAddress

	site := "all"
	if !config.Site.IsNull() && !config.Site.IsUnknown() {
		site = config.Site.ValueString()
	} else if !plan.Site.IsNull() && !plan.Site.IsUnknown() {
		site = plan.Site.ValueString()
	}

	proxyProtocol := false
	if !config.ProxyProtocol.IsNull() && !config.ProxyProtocol.IsUnknown() {
		proxyProtocol = config.ProxyProtocol.ValueBool()
	} else if !plan.ProxyProtocol.IsNull() && !plan.ProxyProtocol.IsUnknown() {
		proxyProtocol = plan.ProxyProtocol.ValueBool()
	}

	endpointReq := mbProxy.EndpointRequest{
		Site:          site,
		ProxyProtocol: proxyProtocol,
	}

	endpoints, err := r.client.Proxy().CreateOrUpdateEndpoints(
		ctx,
		domain,
		hostname,
		address,
		site,
		[]mbProxy.EndpointRequest{endpointReq},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Proxy Endpoint",
			"Could not create Proxy Endpoint, unexpected error: "+err.Error(),
		)
		return
	}

	if len(endpoints) != 1 {
		resp.Diagnostics.AddError(
			"Error creating Proxy Endpoint",
			fmt.Sprintf("Expected 1 endpoint, got %d", len(endpoints)),
		)
		return
	}

	endpoint := endpoints[0]
	state := ProxyEndpointResourceModel{
		ID:            types.StringValue(fmt.Sprintf("%s/%s/%s/%s", endpoint.Domain, endpoint.Hostname, endpoint.Address.String(), endpoint.Site)),
		Domain:        types.StringValue(endpoint.Domain),
		Hostname:      types.StringValue(endpoint.Hostname),
		Address:       types.StringValue(endpoint.Address.String()),
		Site:          types.StringValue(endpoint.Site),
		ProxyProtocol: types.BoolValue(endpoint.ProxyProtocol),
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ProxyEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProxyEndpointResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Domain.IsNull() || state.Domain.IsUnknown() ||
		state.Hostname.IsNull() || state.Hostname.IsUnknown() ||
		state.Address.IsNull() || state.Address.IsUnknown() ||
		state.Site.IsNull() || state.Site.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing proxy endpoint identity",
			"Domain, hostname, address, and site must be set in state to read a proxy endpoint.",
		)
		return
	}

	domain := state.Domain.ValueString()
	hostname := state.Hostname.ValueString()
	address := state.Address.ValueString()
	site := state.Site.ValueString()

	normalizedAddress, err := normalizeIPv6(address)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid IPv6 address in state",
			fmt.Sprintf("Could not parse address %q: %s", address, err.Error()),
		)
		return
	}

	endpoint, found, err := r.client.Proxy().GetEndpoint(ctx, domain, hostname, normalizedAddress, site)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Proxy Endpoint",
			fmt.Sprintf("Could not read Proxy Endpoint %s/%s/%s/%s: %s", domain, hostname, address, site, err.Error()),
		)
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", endpoint.Domain, endpoint.Hostname, endpoint.Address.String(), endpoint.Site))
	state.Domain = types.StringValue(endpoint.Domain)
	state.Hostname = types.StringValue(endpoint.Hostname)
	state.Address = types.StringValue(endpoint.Address.String())
	state.Site = types.StringValue(endpoint.Site)
	state.ProxyProtocol = types.BoolValue(endpoint.ProxyProtocol)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProxyEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProxyEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProxyEndpointResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain
	if domain.IsNull() || domain.IsUnknown() {
		domain = state.Domain
	}
	hostname := plan.Hostname
	if hostname.IsNull() || hostname.IsUnknown() {
		hostname = state.Hostname
	}
	address := plan.Address
	if address.IsNull() || address.IsUnknown() {
		address = state.Address
	}
	site := plan.Site
	if site.IsNull() || site.IsUnknown() {
		site = state.Site
	}

	if domain.IsNull() || domain.IsUnknown() ||
		hostname.IsNull() || hostname.IsUnknown() ||
		address.IsNull() || address.IsUnknown() ||
		site.IsNull() || site.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing proxy endpoint identity",
			"Domain, hostname, address, and site must be set in state to update a proxy endpoint.",
		)
		return
	}

	normalizedAddress, err := normalizeIPv6(address.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("address"), "Invalid IPv6 address", err.Error())
		return
	}

	proxyProtocol := plan.ProxyProtocol
	if proxyProtocol.IsNull() || proxyProtocol.IsUnknown() {
		proxyProtocol = state.ProxyProtocol
	}
	if proxyProtocol.IsNull() || proxyProtocol.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing proxy protocol value",
			"Proxy protocol must be set in state to update a proxy endpoint.",
		)
		return
	}

	endpointReq := mbProxy.EndpointRequest{
		Site:          site.ValueString(),
		ProxyProtocol: proxyProtocol.ValueBool(),
	}

	endpoints, err := r.client.Proxy().CreateOrUpdateEndpoints(
		ctx,
		domain.ValueString(),
		hostname.ValueString(),
		normalizedAddress,
		site.ValueString(),
		[]mbProxy.EndpointRequest{endpointReq},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Proxy Endpoint",
			"Could not update Proxy Endpoint, unexpected error: "+err.Error(),
		)
		return
	}

	if len(endpoints) != 1 {
		resp.Diagnostics.AddError(
			"Error updating Proxy Endpoint",
			fmt.Sprintf("Expected 1 endpoint, got %d", len(endpoints)),
		)
		return
	}

	endpoint := endpoints[0]
	state = ProxyEndpointResourceModel{
		ID:            types.StringValue(fmt.Sprintf("%s/%s/%s/%s", endpoint.Domain, endpoint.Hostname, endpoint.Address.String(), endpoint.Site)),
		Domain:        types.StringValue(endpoint.Domain),
		Hostname:      types.StringValue(endpoint.Hostname),
		Address:       types.StringValue(endpoint.Address.String()),
		Site:          types.StringValue(endpoint.Site),
		ProxyProtocol: types.BoolValue(endpoint.ProxyProtocol),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProxyEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProxyEndpointResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Domain.IsNull() || state.Domain.IsUnknown() ||
		state.Hostname.IsNull() || state.Hostname.IsUnknown() ||
		state.Address.IsNull() || state.Address.IsUnknown() ||
		state.Site.IsNull() || state.Site.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing proxy endpoint identity",
			"Domain, hostname, address, and site must be set in state to delete a proxy endpoint.",
		)
		return
	}

	normalizedAddress, err := normalizeIPv6(state.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid IPv6 address in state",
			fmt.Sprintf("Could not parse address %q: %s", state.Address.ValueString(), err.Error()),
		)
		return
	}

	err = r.client.Proxy().DeleteEndpoints(
		ctx,
		state.Domain.ValueString(),
		state.Hostname.ValueString(),
		normalizedAddress,
		state.Site.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Proxy Endpoint ",
			"Could not delete Proxy Endpoint , unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *ProxyEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected import ID in the format domain/hostname/address/site.",
		)
		return
	}

	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
		if parts[i] == "" {
			resp.Diagnostics.AddError(
				"Invalid import ID",
				"Import ID parts must be non-empty in the format domain/hostname/address/site.",
			)
			return
		}
	}

	normalizedAddress, err := normalizeIPv6(parts[2])
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Address %q is not valid IPv6: %s", parts[2], err.Error()),
		)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(fmt.Sprintf("%s/%s/%s/%s", parts[0], parts[1], normalizedAddress, parts[3])))
	resp.State.SetAttribute(ctx, path.Root("domain"), types.StringValue(parts[0]))
	resp.State.SetAttribute(ctx, path.Root("hostname"), types.StringValue(parts[1]))
	resp.State.SetAttribute(ctx, path.Root("address"), types.StringValue(normalizedAddress))
	resp.State.SetAttribute(ctx, path.Root("site"), types.StringValue(parts[3]))
}
