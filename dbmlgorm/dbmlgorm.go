package dbmlgorm

import (
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/shifty11/dbml-to-gorm/common"
	"github.com/stretchr/stew/slice"
	"path/filepath"
	"regexp"
	"strings"
)

func dbmlToGormString(dbml *core.DBML) string {
	str, types := getProjectConfig(dbml)
	for _, enum := range dbml.Enums {
		str += dbmlEnumToGormString(enum)
	}
	for _, table := range dbml.Tables {
		str += dbmlTableToGormString(table, types)
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

func dbmlTableToGormString(table core.Table, types map[string]string) string {
	str := ""
	str += fmt.Sprintf("type %v struct {\n", table.Name)
	if table.Note != "" {
		str += "\t" + strings.Replace(table.Note, ";", "\n", -1) + "\n"
	}
	for _, column := range table.Columns {
		if column.Settings.Note != "hidden" {
			columnType, isPresent := types[column.Type]
			if !isPresent {
				columnType = column.Type
			}
			settings := createGormSettings(column)
			str += fmt.Sprintf("    %v %v%v\n", column.Name, columnType, settings)
		}
	}
	str += "}\n\n"
	return str
}

func createGormSettings(column core.Column) string {
	re := regexp.MustCompile(`^gorm:\"([a-zA-Z-_;:<> ]*)\"$`)
	var settings []string

	match := re.FindStringSubmatch(column.Settings.Note)
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
		reDefault := regexp.MustCompile(`^gorm:\"[^\"]*?default:[^\"]*?\"$`)
		if len(reDefault.FindString(column.Settings.Note)) == 0 {
			settings = append(settings, "default:"+column.Settings.Default)
		}
	}
	if len(settings) > 0 {
		return fmt.Sprintf(" `gorm:\"%v\"`", strings.Join(settings, ";"))
	}
	return ""
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

func CreateGormFiles(dbml *core.DBML, outputPath string) {
	gormString := dbmlToGormString(dbml)
	common.WriteToFile(gormString, filepath.Join(outputPath, "model.gen.go"))

	migrationString := dbmlToMigrationString(dbml)
	common.WriteToFile(migrationString, filepath.Join(outputPath, "migration.gen.go"))
}
