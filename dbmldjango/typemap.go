package dbmldjango

import "github.com/shifty11/dbml-convert/common"

var types = map[string]string{
	"string":       "models.CharField",
	"uint":         "models.IntegerField",
	"int":          "models.IntegerField",
	"email":        "models.EmailField",
	"datetime":     "models.DateTimeField",
	"nulldatetime": "models.DateTimeField",
	"decimal":      "models.DecimalField",
}

var specialOptions = map[string]string{
	common.OCreatedAt: "auto_now_add=True",
	common.OUpdatedAt: "auto_now=True",
}
