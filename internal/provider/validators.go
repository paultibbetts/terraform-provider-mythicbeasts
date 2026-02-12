// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type multipleOfTenValidator struct{}

func (v multipleOfTenValidator) Description(ctx context.Context) string {
	return "Value must be a multiple of 10"
}

func (v multipleOfTenValidator) MarkdownDescription(ctx context.Context) string {
	return "Value must be a **multiple of 10**"
}

func (v multipleOfTenValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	val := req.ConfigValue.ValueInt64()
	if val%10 != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			fmt.Sprintf("Value %d is not a multiple of 10", val),
		)
	}
}

func MultipleOfTen() validator.Int64 {
	return multipleOfTenValidator{}
}

type multipleOfValidator struct {
	divisor int64
}

func (v multipleOfValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Value must be a multiple of %d", v.divisor)
}

func (v multipleOfValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Value must be a **multiple of %d**", v.divisor)
}

func (v multipleOfValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	val := req.ConfigValue.ValueInt64()
	if val%v.divisor != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			fmt.Sprintf("Value %d is not a multiple of %d", val, v.divisor),
		)
	}
}

func MultipleOf(divisor int64) validator.Int64 {
	return multipleOfValidator{divisor: divisor}
}
