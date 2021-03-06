package dbmldjango

import (
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/gobeam/stringy"
	"github.com/shifty11/dbml-convert/common"
	"github.com/stretchr/stew/slice"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func getTemplate(path string) string {
	template, err := ioutil.ReadFile(path)
	if err != nil {
		return "# Auto generated models. Do not edit by hand!\n# Instead add the file 'models.py.template' " +
			"which will be added to the top of 'models.py'\nimport enum\n\nfrom django.db import models\n\n\n"
	}
	return string(template)
}

func dbmlToDjangoString(pythonFile PythonFile, djangoPath string) string {
	str := getTemplate(filepath.Join(djangoPath, pythonFile.FilePath+".template"))
	for _, enum := range pythonFile.Enums {
		str += dbmlEnumToDjangoString(enum)
	}
	for _, table := range pythonFile.Tables {
		str += dbmlTableToDjangoString(table, pythonFile.Enums)
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
	Meta         []string
}

func parseTableSettings(table core.Table) TableSettings {
	settings := TableSettings{Hidden: false}

	settingsStr := common.GetNoteSettings(table.Note, common.DJangoSettings)

	for _, entry := range settingsStr {
		if entry == common.SHidden {
			return TableSettings{Hidden: true}
		} else if strings.HasPrefix(entry, "inherit=") {
			for _, inhStr := range strings.Split(strings.Replace(entry, "inherit=", "", 1), ";") {
				str := stringy.New(strings.Replace(inhStr, ".", "", -1)).CamelCase("?", "")
				settings.Inheritances = append(settings.Inheritances, str)
			}
		} else if strings.HasPrefix(entry, "model_path=") {
			settings.ModelPath = strings.Replace(entry, "model_path=", "", 1)
		} else if strings.HasPrefix(entry, "meta=") {
			meta := strings.Replace(entry[:len(entry)-1], "meta=[", "", 1)
			settings.Meta = strings.Split(meta, " ")
		}
	}
	if settings.ModelPath == "" {
		panic(fmt.Sprintf("Table %v has no model_path", table.Name))
	}
	return settings
}

func parseColumnParameters(column core.Column) []string {
	var params []string

	settingsStr := common.GetNoteSettings(column.Settings.Note, common.DJangoSettings)
	for _, settings := range settingsStr {
		for _, entry := range strings.Split(settings, ";") {
			params = append(params, entry)
		}
	}
	params = append(params, getDbmlColumnSettings(column)...)
	return params
}

func getDbmlColumnSettings(column core.Column) []string {
	var settings []string

	key := "unique"
	if column.Settings.Unique && !slice.Contains(settings, key) {
		settings = append(settings, "unique=True")
	}
	key = "not null"
	if column.Settings.Null && !slice.Contains(settings, key) {
		settings = append(settings, "null=True")
	}
	key = "default"
	if column.Settings.Default != "" && !slice.Contains(settings, key) {
		settings = append(settings, fmt.Sprintf("default=%v", parseDefault(column)))
	}
	return settings
}

func parseDefault(column core.Column) string {
	if column.Type == common.TBool || column.Type == common.TBoolean {
		return strings.Title(column.Settings.Default)
	}
	return fmt.Sprintf("'%v'", column.Settings.Default)
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
	if settings.Hidden {
		return ""
	}
	inheritance := "models.Model"
	if len(settings.Inheritances) > 0 {
		inheritance = strings.Join(settings.Inheritances, ", ")
	}
	str += fmt.Sprintf("class %v(%v):\n", table.Name, inheritance)
	for _, column := range table.Columns {
		columnSettings := common.GetNoteSettings(column.Settings.Note, common.DJangoSettings)
		if !slice.Contains(columnSettings, common.SHidden) {
			columnType := types[column.Type]
			columnParams := parseColumnParameters(column)
			paramsString := ""
			if len(columnParams) > 0 {
				if columnParams[0] == common.OCreatedAt {
					paramsString = specialOptions[common.OCreatedAt]
				} else if columnParams[0] == common.OUpdatedAt {
					paramsString = specialOptions[common.OUpdatedAt]
				} else {
					paramsString = strings.Join(columnParams, ", ")
				}
			}
			if columnType == "" {
				if strings.HasPrefix(column.Type, "[]") {
					continue
				}
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
	if len(settings.Meta) > 0 {
		str += "\n    class Meta:\n"
		for _, entry := range settings.Meta {
			str += fmt.Sprintf("        %v\n", strings.Replace(entry, "=", " = ", 1))
		}
		str += "\n\n"
	} else {
		str += fmt.Sprintf("\n    class Meta:\n        db_table = '%vs'\n\n\n", tableName)
	}
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
	files := map[string]*PythonFile{}
	for _, table := range dbml.Tables {
		settings := parseTableSettings(table)
		if !settings.Hidden {
			djangoTable := DjangoTable{Table: table, Settings: settings}
			if file, ok := files[settings.ModelPath]; ok {
				file.Tables = append(file.Tables, djangoTable)
				file.Enums = addEnums(file.Enums, dbml.Enums, table)
			} else {
				files[settings.ModelPath] = &PythonFile{
					FilePath: settings.ModelPath,
					Tables:   []DjangoTable{djangoTable},
					Enums:    addEnums([]core.Enum{}, dbml.Enums, table),
				}
			}
		}
	}
	dbmlDjango := DBMLDjango{}
	for _, file := range files {
		dbmlDjango.Files = append(dbmlDjango.Files, *file)
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
