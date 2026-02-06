// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func normalizeIPv6(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	addr, err := netip.ParseAddr(trimmed)
	if err != nil {
		return "", err
	}
	if !addr.Is6() {
		return "", fmt.Errorf("address %q is not IPv6", trimmed)
	}
	return addr.String(), nil
}

type ipv6NormalizePlanModifier struct{}

func (m ipv6NormalizePlanModifier) Description(_ context.Context) string {
	return "Normalize IPv6 addresses to canonical form."
}

func (m ipv6NormalizePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Normalize IPv6 addresses to canonical form."
}

func (m ipv6NormalizePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	normalized, err := normalizeIPv6(req.ConfigValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid IPv6 address", err.Error())
		return
	}

	resp.PlanValue = types.StringValue(normalized)
}

func IPv6Normalize() planmodifier.String {
	return ipv6NormalizePlanModifier{}
}
