package dbmlgorm

import (
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/duythinht/dbml-go/parser"
	"github.com/duythinht/dbml-go/scanner"
	"log"
	"os"
)

func dbmlToGormString(dbml *core.DBML) string {
	str, types := getProjectConfig(dbml)
	for _, table := range dbml.Tables {
		str += dbmlTableToGormString(table, types)
	}
	return str
}

func dbmlTableToGormString(table core.Table, types map[string]string) string {
	str := ""
	str += fmt.Sprintf("type %v struct {\n", table.Name)
	for _, column := range table.Columns {
		columnType, isPresent := types[column.Type]
		if !isPresent {
			columnType = column.Type
		}
		str += fmt.Sprintf("    %v %v", column.Name, columnType)
		if column.Settings.Note != "" {
			str += fmt.Sprintf(" `%v`", column.Settings.Note)
		}
		str += "\n"
	}
	str += "}\n\n"
	return str
}

func WriteToGormFile(gormString string, outputPath string) {
	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.WriteString(gormString)
	if err != nil {
		log.Fatal(err)
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
}

func ParseDbml(dbmlPath string) string {
	file, err := os.Open(dbmlPath)
	if err != nil {
		log.Fatal(err)
	}

	scan := scanner.NewScanner(file)
	pars := parser.NewParser(scan)
	dbml, err := pars.Parse()
	if err != nil {
		log.Fatal(err)
	}

	return dbmlToGormString(dbml)
}
