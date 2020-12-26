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
	str := "// Auto-generated code\n\npackage model\n\n"
	for _, table := range dbml.Tables {
		str += dbmlTableToGormString(table)
	}
	return str
}

func dbmlTableToGormString(table core.Table) string {
	str := ""
	str += fmt.Sprintf("type %v struct {\n", table.Name)
	for _, column := range table.Columns {
		str += fmt.Sprintf("    %v %v", column.Name, column.Type)
		if column.Settings.Note != "" {
			str += fmt.Sprintf(" `%v`", column.Settings.Note)
		}
		str += "\n"
	}
	str += "}\n\n"
	return str
}

func WriteToGormFile(gormString string, outputPath string) {
	file, err1 := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err1 != nil {
		log.Fatal(err1)
	}
	_, err2 := file.WriteString(gormString)
	if err2 != nil {
		log.Fatal(err2)
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
