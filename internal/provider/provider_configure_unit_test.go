// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/paultibbetts/mythicbeasts-client-go"
)

func TestProviderConfigureMissingCredentials(t *testing.T) {
	t.Setenv("MYTHICBEASTS_KEYID", "")
	t.Setenv("MYTHICBEASTS_SECRET", "")

	p := &mythicbeastsProvider{version: "test"}
	req := fwprovider.ConfigureRequest{
		Config: testProviderConfig(testProviderSchema(t), nil, nil),
	}
	resp := &fwprovider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics errors when credentials are missing")
	}

	if resp.Diagnostics.ErrorsCount() != 2 {
		t.Fatalf("expected 2 diagnostics errors, got %d", resp.Diagnostics.ErrorsCount())
	}
}

func TestProviderConfigureUsesEnvironmentCredentials(t *testing.T) {
	t.Setenv("MYTHICBEASTS_KEYID", "env-key")
	t.Setenv("MYTHICBEASTS_SECRET", "env-secret")

	p := &mythicbeastsProvider{version: "test"}
	req := fwprovider.ConfigureRequest{
		Config: testProviderConfig(testProviderSchema(t), nil, nil),
	}
	resp := &fwprovider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no diagnostics errors, got %d", resp.Diagnostics.ErrorsCount())
	}

	dsClient, ok := resp.DataSourceData.(*mythicbeasts.Client)
	if !ok {
		t.Fatalf("expected DataSourceData to be *mythicbeasts.Client, got %T", resp.DataSourceData)
	}

	rsClient, ok := resp.ResourceData.(*mythicbeasts.Client)
	if !ok {
		t.Fatalf("expected ResourceData to be *mythicbeasts.Client, got %T", resp.ResourceData)
	}

	if dsClient != rsClient {
		t.Fatalf("expected DataSourceData and ResourceData to reference the same client instance")
	}

	if dsClient.Auth.KeyID != "env-key" {
		t.Fatalf("expected client keyid to come from environment, got %q", dsClient.Auth.KeyID)
	}

	if dsClient.Auth.Secret != "env-secret" {
		t.Fatalf("expected client secret to come from environment, got %q", dsClient.Auth.Secret)
	}
}

func TestProviderConfigureConfigOverridesEnvironmentCredentials(t *testing.T) {
	t.Setenv("MYTHICBEASTS_KEYID", "env-key")
	t.Setenv("MYTHICBEASTS_SECRET", "env-secret")

	configKey := "config-key"
	configSecret := "config-secret"

	p := &mythicbeastsProvider{version: "test"}
	req := fwprovider.ConfigureRequest{
		Config: testProviderConfig(testProviderSchema(t), &configKey, &configSecret),
	}
	resp := &fwprovider.ConfigureResponse{}

	p.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no diagnostics errors, got %d", resp.Diagnostics.ErrorsCount())
	}

	dsClient, ok := resp.DataSourceData.(*mythicbeasts.Client)
	if !ok {
		t.Fatalf("expected DataSourceData to be *mythicbeasts.Client, got %T", resp.DataSourceData)
	}

	if dsClient.Auth.KeyID != configKey {
		t.Fatalf("expected config keyid to override environment, got %q", dsClient.Auth.KeyID)
	}

	if dsClient.Auth.Secret != configSecret {
		t.Fatalf("expected config secret to override environment, got %q", dsClient.Auth.Secret)
	}
}

func testProviderSchema(t *testing.T) providerschema.Schema {
	t.Helper()

	p := &mythicbeastsProvider{version: "test"}
	resp := &fwprovider.SchemaResponse{}
	p.Schema(context.Background(), fwprovider.SchemaRequest{}, resp)

	return resp.Schema
}

func testProviderConfig(s providerschema.Schema, keyid, secret *string) tfsdk.Config {
	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"keyid":  tftypes.String,
			"secret": tftypes.String,
		},
	}

	keyidValue := tftypes.NewValue(tftypes.String, nil)
	if keyid != nil {
		keyidValue = tftypes.NewValue(tftypes.String, *keyid)
	}

	secretValue := tftypes.NewValue(tftypes.String, nil)
	if secret != nil {
		secretValue = tftypes.NewValue(tftypes.String, *secret)
	}

	return tfsdk.Config{
		Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
			"keyid":  keyidValue,
			"secret": secretValue,
		}),
		Schema: s,
	}
}
