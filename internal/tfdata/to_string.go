package tfdata

import (
	"fmt"

	"github.com/jeremmfr/junosdecode"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func JunosDecode(str, errMsg string) (basetypes.StringValue, error) {
	v, err := junosdecode.Decode(str)
	if err != nil {
		return types.StringNull(), fmt.Errorf("decoding "+errMsg+": %w", err)
	}

	return types.StringValue(v), nil
}
