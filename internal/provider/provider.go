// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/paultibbetts/mythicbeasts-client-go"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &mythicbeastsProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mythicbeastsProvider{
			version: version,
		}
	}
}

// mythicbeastsProvider is the provider implementation.
type mythicbeastsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type mythicbeastsProviderModel struct {
	KeyID  types.String `tfsdk:"keyid"`
	Secret types.String `tfsdk:"secret"`
}

// Metadata returns the provider type name.
func (p *mythicbeastsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mythicbeasts"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *mythicbeastsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"keyid": schema.StringAttribute{
				Optional: true,
			},
			"secret": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a mythicbeasts API client for data sources and resources.
func (p *mythicbeastsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config mythicbeastsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.KeyID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("keyid"),
			"Unknown Mythic Beasts API Key",
			"The provider cannot create the Mythic Beasts API client as there is an unknown configuration value for the Mythic Beasts API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MYTHICBEASTS_KEY environment variable.",
		)
	}

	if config.Secret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Unknown Mythic Beasts API secret",
			"The provider cannot create the Mythic Beasts API client as there is an unknown configuration value for the Mythic Beasts API secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MYTHICBEASTS_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	keyid := os.Getenv("MYTHICBEASTS_KEYID")
	secret := os.Getenv("MYTHICBEASTS_SECRET")

	if !config.KeyID.IsNull() {
		keyid = config.KeyID.ValueString()
	}

	if !config.Secret.IsNull() {
		secret = config.Secret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if keyid == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("keyid"),
			"Missing Mythic Beasts API keyid",
			"The provider cannot create the Mythic Beasts API client as there is a missing or empty value for the Mythic Beasts API keyid. "+
				"Set the key value in the configuration or use the MYTHICBEASTS_KEYID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if secret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Missing Mythic Beasts API secret",
			"The provider cannot create the Mythic Beasts API client as there is a missing or empty value for the Mythic Beasts API secret. "+
				"Set the secret value in the configuration or use the MYTHICBEATS_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new mythicbeasts client using the configuration values
	client, err := mythicbeasts.NewClient(keyid, secret)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Mythic Beasts API Client",
			"An unexpected error occurred when creating the Mythic Beasts API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Mythic Beasts Client Error: "+err.Error(),
		)
		return
	}

	// Make the mythicbeasts client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *mythicbeastsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPiModelsDataSource,
		NewPiOperatingSystemsDataSource,
		NewVPSDiskSizesDataSource,
		NewVPSHostsDataSource,
		NewVPSImagesDataSource,
		NewVPSPricingDataSource,
		NewVPSProductsDataSource,
		NewVPSZonesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *mythicbeastsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPiResource,
		NewProxyEndpointResource,
		NewUserDataResource,
		NewVPSResource,
	}
}
