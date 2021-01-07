package dbmldjango

var types = map[string]string{
	"string":       "models.CharField",
	"uint":         "models.IntegerField",
	"int":          "models.IntegerField",
	"email":        "models.EmailField",
	"datetime":     "models.DateTimeField",
	"nulldatetime": "models.DateTimeField",
	"decimal":      "models.DecimalField",
}
