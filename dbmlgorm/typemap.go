package dbmlgorm

var types = map[string]string{
	"string":       "string",
	"uint":         "uint",
	"int":          "int",
	"email":        "string",
	"datetime":     "time.Time",
	"nulldatetime": "NullTime",
	"decimal":      "decimal.Decimal",
}
