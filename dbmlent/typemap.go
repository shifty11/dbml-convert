package dbmlent

import "github.com/shifty11/dbml-convert/common"

var typeMap = map[string]string{
	common.TString:   "field.String",
	common.TInt:      "field.Int",
	common.TUint:     "field.Int",
	common.TEmail:    "field.String",
	common.TDatetime: "field.Time",
	common.TDecimal:  "field.String",
}

var specialOptions = map[string]string{
	common.OCreatedAt: "Default(time.Now)",
	common.OUpdatedAt: "Default(time.Now).\\nUpdateDefault(time.Now)",
}
