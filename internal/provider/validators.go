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
