package dbmlgorm

import (
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/duythinht/dbml-go/parser"
	"github.com/duythinht/dbml-go/scanner"
	"log"
	"os"
	"path/filepath"
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

func writeToGormFile(gormString string, outputPath string) {
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

var migrationTemplate = `import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect(dsn string) *gorm.DB {
	if db == nil {
		newDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		db = newDb
	}
	return db
}

func Migrate(dsn string) {
	db := Connect(dsn)
`

func dbmlToMigrationString(dbml *core.DBML) string {
	migrationStr := getProjectHeader(dbml)
	migrationStr += migrationTemplate
	for _, table := range dbml.Tables {
		migrationStr += fmt.Sprintf(`
	if err := db.AutoMigrate(&%v{}); err != nil {
		panic("failed to migrate %v")
	}`, table.Name, table.Name)
	}
	migrationStr += `
	fmt.Println("Migration succeeded")
}`
	return migrationStr
}

func CreateGormFiles(dbmlPath string, outputPath string) {
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

	gormString := dbmlToGormString(dbml)
	writeToGormFile(gormString, filepath.Join(outputPath, "model.gen.go"))

	migrationString := dbmlToMigrationString(dbml)
	writeToGormFile(migrationString, filepath.Join(outputPath, "migration.gen.go"))
}
