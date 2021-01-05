package dbmldjango

import (
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/gobeam/stringy"
	"github.com/shifty11/dbml-to-gorm/common"
	"github.com/stretchr/stew/slice"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var djangoRe = regexp.MustCompile(`django:\"([a-zA-Z0-9-_;:<>= \.\/\']*)\"`)

func getTemplate(path string) string {
	template, err := ioutil.ReadFile(path)
	if err != nil {
		return "# Auto generated models. Do not edit by hand!\n# Instead add the file 'models.py.template' " +
			"which will be added to the top of 'models.py'\nimport enum\n\nfrom django.db import models\n\n\n"
	}
	return string(template)
}

func dbmlToDjangoString(dbml PythonFile, djangoPath string) string {
	str := getTemplate(djangoPath + "models.py.template")
	for _, enum := range dbml.Enums {
		str += dbmlEnumToDjangoString(enum)
	}
	for _, table := range dbml.Tables {
		str += dbmlTableToDjangoString(table, dbml.Enums)
	}
	return str
}

func dbmlEnumToDjangoString(enum core.Enum) string {
	str := ""
	str += fmt.Sprintf("class %v(enum.IntEnum):\n", enum.Name)
	for i, column := range enum.Values {
		str += fmt.Sprintf("    %v = %v", column.Name, i)
		if column.Note != "" {
			str += fmt.Sprintf("\t# %v", column.Note)
		}
		str += "\n"
	}
	str += "\n\n"
	return str
}

type TableSettings struct {
	Inheritances []string
	ModelPath    string
	Hidden       bool
}

func parseTableSettings(table core.Table) TableSettings {
	settings := TableSettings{Hidden: false}

	match := djangoRe.FindStringSubmatch(table.Note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], ";") {
			if entry == "hidden" {
				return TableSettings{Hidden: true}
			} else if strings.HasPrefix(entry, "inherit=") {
				for _, inhStr := range strings.Split(strings.Replace(entry, "inherit=", "", 1), ";") {
					str := stringy.New(strings.Replace(inhStr, ".", "", -1)).CamelCase("?", "")
					settings.Inheritances = append(settings.Inheritances, str)
				}
			} else if strings.HasPrefix(entry, "model_path=") {
				settings.ModelPath = strings.Replace(entry, "model_path=", "", 1)
			}
		}
	}
	if settings.ModelPath == "" {
		panic(fmt.Sprintf("Table %v has no model_path", table.Name))
	}
	return settings
}

func parseColumnParameters(column core.Column) []string {
	params := getDbmlColumnSettings(column)

	match := djangoRe.FindStringSubmatch(column.Settings.Note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], ";") {
			params = append(params, entry)
		}
	}
	return params
}

func getDbmlColumnSettings(column core.Column) []string {
	var settings []string

	key := "unique"
	if column.Settings.Unique && !slice.Contains(settings, key) {
		settings = append(settings, "unique=True")
	}
	key = "not null"
	if !column.Settings.Null && !slice.Contains(settings, key) {
		settings = append(settings, "null=True")
	}
	key = "default"
	if column.Settings.Default != "" && !slice.Contains(settings, key) {
		settings = append(settings, fmt.Sprintf("default='%v'", column.Settings.Default))
	}
	return settings
}

func getRelationType(column core.Column, paramsString string) string {
	columnType := ""
	if len(paramsString) > 0 {
		paramsString = ", " + paramsString
	}
	if column.Settings.Ref.Type == core.OneToOne {
		columnType = "models.OneToOneField"
	} else if column.Settings.Ref.Type == core.ManyToOne || column.Settings.Ref.Type == core.OneToMany {
		columnType = "models.ForeignKey"
	} else {
		fmt.Printf("%v was not found in type definition\n", column.Type)
		return ""
	}
	return columnType
}

func getEnumType(enums []core.Enum, column core.Column, columnType string, paramsString string) (string, string) {
	for _, enum := range enums {
		if column.Type == enum.Name {
			columnType = "models.CharField"
			if len(paramsString) > 0 {
				paramsString = paramsString + ", "
			}
			paramsString += fmt.Sprintf("choices=[(tag, tag.value) for tag in %v]", column.Type)
		}
	}
	return columnType, paramsString
}

func dbmlTableToDjangoString(djangoTable DjangoTable, enums []core.Enum) string {
	str := ""
	table := djangoTable.Table
	settings := parseTableSettings(table)
	inheritance := "models.Model"
	if len(settings.Inheritances) > 0 {
		inheritance = strings.Join(settings.Inheritances, ", ")
	}
	str += fmt.Sprintf("class %v(%v):\n", table.Name, inheritance)
	for _, column := range table.Columns {
		if column.Settings.Note != "hidden" {
			columnType := types[column.Type]
			columnParams := parseColumnParameters(column)
			paramsString := ""
			if len(columnParams) > 0 {
				paramsString = strings.Join(columnParams, ", ")
			}
			if columnType == "" {
				if column.Settings.Ref.Type != core.None {
					columnType = getRelationType(column, paramsString)
					if columnType == "" && paramsString == "" {
						continue
					}
				} else {
					columnType, paramsString = getEnumType(enums, column, columnType, paramsString)
				}
			}
			columnName := strings.ToLower(stringy.New(column.Name).SnakeCase("?", "").Get())
			str += fmt.Sprintf("    %v = %v(%v)\n", columnName, columnType, paramsString)
		}
	}
	tableName := strings.ToLower(stringy.New(table.Name).SnakeCase("?", "").Get())
	str += fmt.Sprintf("\n    class Meta:\n        db_table = '%vs'\n\n", tableName)
	return str
}

type DBMLDjango struct {
	Files []PythonFile
}

type PythonFile struct {
	FilePath string
	Tables   []DjangoTable
	Enums    []core.Enum
}

type DjangoTable struct {
	Table    core.Table
	Settings TableSettings
}

func addEnums(currentEnums []core.Enum, enums []core.Enum, table core.Table) []core.Enum {
	for _, enum := range enums {
		hasEnum := false
		for _, currentEnum := range currentEnums {
			if currentEnum.Name == enum.Name {
				hasEnum = true
			}
		}
		if hasEnum {
			continue
		}
		for _, column := range table.Columns {
			if enum.Name == column.Type {
				currentEnums = append(currentEnums, enum)
			}
		}
	}
	return currentEnums
}

func dbmlSplitByModelPath(dbml *core.DBML) DBMLDjango {
	files := map[string]PythonFile{}
	for _, table := range dbml.Tables {
		settings := parseTableSettings(table)
		if !settings.Hidden {
			djangoTable := DjangoTable{Table: table, Settings: settings}
			if file, ok := files[settings.ModelPath]; ok {
				file.Tables = append(file.Tables, djangoTable)
				file.Enums = addEnums(file.Enums, dbml.Enums, table)
			} else {
				files[settings.ModelPath] = PythonFile{
					FilePath: settings.ModelPath,
					Tables:   []DjangoTable{djangoTable},
					Enums:    addEnums([]core.Enum{}, dbml.Enums, table),
				}
			}
		}
	}
	dbmlDjango := DBMLDjango{}
	for _, file := range files {
		dbmlDjango.Files = append(dbmlDjango.Files, file)
	}
	return dbmlDjango
}

func CreateDjangoFiles(dbml *core.DBML, djangoRoot string) {
	dbmlDjango := dbmlSplitByModelPath(dbml)

	for _, file := range dbmlDjango.Files {
		djangoString := dbmlToDjangoString(file, djangoRoot)
		common.WriteToFile(djangoString, filepath.Join(djangoRoot, file.FilePath))
	}
}
