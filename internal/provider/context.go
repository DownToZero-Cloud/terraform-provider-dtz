package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DtzContext struct {
	Id    types.String `tfsdk:"id"`
	Alias types.String `tfsdk:"alias"`
}
