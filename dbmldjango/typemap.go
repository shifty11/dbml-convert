package dbmldjango

import "github.com/shifty11/dbml-convert/common"

var types = map[string]string{
	common.TString:   "models.CharField",
	common.TUint:     "models.IntegerField",
	common.TInt:      "models.IntegerField",
	common.TEmail:    "models.EmailField",
	common.TDatetime: "models.DateTimeField",
	//"nulldatetime": "models.DateTimeField",
	common.TDecimal: "models.DecimalField",
	common.TBool:    "models.BooleanField",
	common.TBoolean: "models.BooleanField",
}

var specialOptions = map[string]string{
	common.OCreatedAt: "auto_now_add=True",
	common.OUpdatedAt: "auto_now=True",
}
