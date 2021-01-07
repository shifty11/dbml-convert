package dbmlgorm

import (
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/shifty11/dbml-convert/common"
	"github.com/stretchr/stew/slice"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var gormRe = regexp.MustCompile(`gorm:"([a-zA-Z0-9-_;:<>= ./']*)"`)

func getTemplate(path string) string {
	template, err := ioutil.ReadFile(path)
	if err != nil {
		return "// Auto generated models. Do not edit by hand!\n// Instead add the file 'model.go.template' " +
			"which will be added to the top of 'model.go'\nimport enum\n\nfrom django.db import models\n\n\n"
	}
	return string(template)
}

func dbmlToGormString(dbml *core.DBML, outputPath string) string {
	str := getTemplate(filepath.Join(outputPath, "model.go.template"))
	for _, enum := range dbml.Enums {
		str += dbmlEnumToGormString(enum)
	}
	for _, table := range dbml.Tables {
		str += dbmlTableToGormString(table)
	}
	return str
}

func dbmlEnumToGormString(enum core.Enum) string {
	str := ""
	str += fmt.Sprintf("type %v uint\n\nconst (\n", enum.Name)
	for i, column := range enum.Values {
		str += fmt.Sprintf("    %v", column.Name)
		if i == 0 {
			str += fmt.Sprintf(" = iota")
		}
		if column.Note != "" {
			str += fmt.Sprintf("\t// %v", column.Note)
		}
		str += "\n"
	}
	str += ")\n\n"
	return str
}

type TableSettings struct {
	Inheritances []string
	Hidden       bool
}

func parseTableSettings(table core.Table) TableSettings {
	settings := TableSettings{Hidden: false}

	match := gormRe.FindStringSubmatch(table.Note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], ";") {
			if entry == "hidden" {
				return TableSettings{Hidden: true}
			} else if strings.HasPrefix(entry, "inherit=") {
				for _, inhStr := range strings.Split(strings.Replace(entry, "inherit=", "", 1), ";") {
					settings.Inheritances = append(settings.Inheritances, inhStr)
				}
			}
		}
	}
	return settings
}

func dbmlTableToGormString(table core.Table) string {
	str := ""
	settings := parseTableSettings(table)
	if settings.Hidden {
		return ""
	}
	str += fmt.Sprintf("type %v struct {\n", table.Name)
	if len(settings.Inheritances) > 0 {
		for _, inheritance := range settings.Inheritances {
			str += fmt.Sprintf("    %v\n", inheritance)
		}
	}
	for _, column := range table.Columns {
		if column.Settings.Note != "hidden" {
			columnType := types[column.Type]
			columnParams := parseColumnParameters(column)

			if columnType == "" {
				if column.Settings.Ref.Type != core.None {
					if column.Settings.Ref.Type == core.ManyToOne {
						str += fmt.Sprintf("    %v int\n", column.Name+"ID")
						columnType = column.Type
					} else {
						panic(fmt.Sprintf("Relation type %v is not yet supported", column.Settings.Ref.Type))
					}
				} else {
					columnType = column.Type
				}
			}
			str += fmt.Sprintf("    %v %v%v\n", column.Name, columnType, columnParams)
		}
	}
	str += "}\n\n"
	return str
}

func parseColumnParameters(column core.Column) string {
	var settings []string

	match := gormRe.FindStringSubmatch(column.Settings.Note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], ";") {
			settings = append(settings, strings.ToLower(entry))
		}
	}

	primarykey := "primarykey"
	if column.Settings.PK && !slice.Contains(settings, primarykey) {
		settings = append(settings, primarykey)
	}
	key := "autoincrement"
	if column.Settings.Increment && !slice.Contains(settings, key) && !slice.Contains(settings, primarykey) {
		settings = append(settings, key) // Add auto-increment just if not primarykey
	}
	key = "unique"
	if column.Settings.Unique && !slice.Contains(settings, key) && !slice.Contains(settings, primarykey) {
		settings = append(settings, key)
	}
	key = "not null"
	if !column.Settings.Null && !slice.Contains(settings, key) && !slice.Contains(settings, primarykey) {
		settings = append(settings, key)
	}
	key = "default"
	if column.Settings.Default != "" && !slice.Contains(settings, key) {
		reDefault := regexp.MustCompile(`^gorm:"[^"]*?default:[^"]*?"$`)
		if len(reDefault.FindString(column.Settings.Note)) == 0 {
			settings = append(settings, "default:"+column.Settings.Default)
		}
	}
	if len(settings) > 0 {
		return fmt.Sprintf(" `gorm:\"%v\"`", strings.Join(settings, ";"))
	}
	return ""
}

func CreateGormFiles(dbml *core.DBML, outputPath string) {
	gormString := dbmlToGormString(dbml, outputPath)
	common.WriteToFile(gormString, filepath.Join(outputPath, "model.gen.go"))
}
