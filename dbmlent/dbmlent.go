package dbmlent

import (
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/gobeam/stringy"
	"github.com/shifty11/dbml-convert/common"
	"github.com/stretchr/stew/slice"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var entRe = regexp.MustCompile(`ent:"([a-zA-Z0-9-_;:<>= ./']*)"`)

type TableSettings struct {
	Hidden bool
}

func parseTableSettings(table core.Table) TableSettings {
	settings := TableSettings{Hidden: false}

	match := entRe.FindStringSubmatch(table.Note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], ";") {
			if entry == common.SHidden {
				return TableSettings{Hidden: true}
			}
		}
	}
	return settings
}

func getImport(table core.Table, fieldsString string) string {
	var imports = []string{"github.com/facebook/ent", "github.com/facebook/ent/schema/field"}
	hasEdges := false
	for _, column := range table.Columns {
		if column.Settings.Note != common.SHidden {
			if column.Settings.Ref.Type == core.None {
				hasEdges = true
				break
			}
		}
	}
	if hasEdges {
		imports = append(imports, "github.com/facebook/ent/schema/edge")
	}
	if hasDecimal(table) {
		imports = append(imports, "github.com/shopspring/decimal")
	}
	if strings.Contains(fieldsString, "time.") {
		imports = append(imports, "time")
	}
	if len(imports) == 1 {
		return fmt.Sprintf("import %v", imports[0])
	}
	str := "import (\n"
	sort.Strings(imports)
	for _, imp := range imports {
		str += "\t\"" + imp + "\"\n"
	}
	str += ")\n"
	return str
}

func hasDecimal(table core.Table) bool {
	for _, column := range table.Columns {
		if column.Settings.Note != common.SHidden {
			if column.Type == common.TDecimal {
				return true
			}
		}
	}
	return false
}

func getSpecialDeclarations(table core.Table) string {
	if hasDecimal(table) {
		return "var dec decimal.Decimal"
	}
	return ""
}

func getFields(table core.Table, dbml *core.DBML) string {
	if len(table.Columns) == 0 {
		return "nil"
	}
	fields := "[]ent.Field{\n"
	for _, column := range table.Columns {
		settings := common.GetNoteSettings(column.Settings.Note, common.EntSettings)
		if !slice.Contains(settings, common.SHidden) &&
			!slice.Contains(settings, common.SBackref) &&
			column.Settings.Ref.Type == core.None {
			columnType := typeMap[column.Type]
			if columnType == "" {
				fields += getEnumField(column, dbml)
			} else {
				columnName := strings.ToLower(stringy.New(column.Name).SnakeCase("?", "").Get())
				fields += fmt.Sprintf("\t\t%v(\"%v\")%v%v,\n",
					columnType, columnName, getFieldExtras(column), formatSettings(settings))
			}
		}
	}
	fields += "\t}"
	return fields
}

func formatSettings(settings []string) string {
	if len(settings) > 0 {
		str := "."
		for _, s := range settings {
			str += "\n\t\t\t" + strings.ReplaceAll(s, "\\n", "\n\t\t\t")
		}
		return str
	}
	return ""
}

func getEnumField(column core.Column, dbml *core.DBML) string {
	for _, enum := range dbml.Enums {
		if strings.ToLower(enum.Name) == strings.ToLower(column.Type) {
			columnName := strings.ToLower(stringy.New(enum.Name).SnakeCase("?", "").Get())
			var enumValues []string
			for _, value := range enum.Values {
				enumValues = append(enumValues, `"`+value.Name+`"`)
			}
			valuesStr := fmt.Sprintf("Values(%v)", strings.Join(enumValues, ","))
			return fmt.Sprintf("\t\tfield.Enum(\"%v\").\n\t\t\t%v%v,\n", columnName, valuesStr, getFieldExtras(column))
		}
	}
	panic(fmt.Sprintf("unknown field type: %v", column.Type))
}

func getFieldExtras(column core.Column) string {
	extras := ""
	if column.Type == common.TDecimal {
		extras += ".\n\t\t\tGoType(&dec)"
	}
	if column.Settings.Null {
		extras += ".\n\t\t\tOptional()"
	}
	if column.Settings.Unique {
		extras += ".\n\t\t\tUnique()"
	}
	if column.Settings.Default != "" {
		extras += ".\n\t\t\tDefault(" + column.Settings.Default + ")"
	}
	return extras
}

func getEdges(table core.Table) string {
	var edges []string
	for _, column := range table.Columns {
		if column.Settings.Note != common.SHidden {
			if column.Settings.Ref.Type == core.ManyToOne {
				split := strings.Split(column.Settings.Ref.To, ".")
				ref := strings.ToLower(stringy.New(split[len(split)-1]).SnakeCase("?", "").Get())
				fromName := strings.ToLower(stringy.New(column.Name).SnakeCase("?", "").Get())
				edges = append(edges, fmt.Sprintf(edgeTemplateFrom, fromName, column.Type, ref))
			} else {
				params := common.GetNoteSettings(column.Settings.Note, common.EntSettings)
				if slice.Contains(params, common.SBackref) {
					toName := strings.ToLower(stringy.New(column.Name).SnakeCase("?", "").Get())
					edgeType := column.Type
					if strings.HasPrefix(edgeType, "[]") {
						edgeType = edgeType[2:]
					}
					ref := strings.ToLower(stringy.New(table.Name).SnakeCase("?", "").Get()) + "_id"
					edges = append(edges, fmt.Sprintf(edgeTemplateTo, toName, edgeType, ref))
				}
			}
		}
	}
	if len(edges) == 0 {
		return "nil"
	}
	str := "[]ent.Edge{"
	for _, edge := range edges {
		str += edge
	}
	str += "\t}"
	return str
}

func dbmlTableToEntString(table core.Table, dbml *core.DBML) string {
	settings := parseTableSettings(table)
	if settings.Hidden {
		return ""
	}
	fields := getFields(table, dbml)
	str := fmt.Sprintf(entTemplate, getImport(table, fields),
		table.Name, table.Name, table.Name,
		getSpecialDeclarations(table),
		table.Name, table.Name,
		fields,
		table.Name, table.Name,
		getEdges(table),
	)
	return str
}

func CreateEntFiles(dbml *core.DBML, outputPath string) {
	for _, table := range dbml.Tables {
		str := dbmlTableToEntString(table, dbml)
		common.WriteToFile(str, filepath.Join(outputPath, strings.ToLower(table.Name)+".go"))
	}
}
