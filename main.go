package main

import (
	"flag"
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/duythinht/dbml-go/parser"
	"github.com/duythinht/dbml-go/scanner"
	"github.com/shifty11/dbml-convert/dbmldjango"
	"github.com/shifty11/dbml-convert/dbmlent"
	"github.com/shifty11/dbml-convert/dbmlgorm"
	"os"
)

func parseArgs() (string, string, bool, bool, bool) {
	flag.Usage = func() { // Showing useful information when the user enters the --help option
		flag.PrintDefaults()
		fmt.Printf("-django|-gorm|-ent <path-to-dbml-file> <path-to-output>\n")
	}
	toDjango := flag.Bool("django", false, "Creates Django models")
	toGorm := flag.Bool("gorm", false, "Creates Gorm models")
	toEnt := flag.Bool("ent", false, "Creates Ent models")
	flag.Parse()

	if len(flag.Args()) < 2 || (!*toDjango && !*toGorm && !*toEnt) || (*toDjango && *toGorm && *toEnt) {
		flag.Usage()
		os.Exit(1)
	}

	dbmlPath := flag.Arg(0)
	outputPath := flag.Arg(1)

	if dbmlPath == "" || outputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return dbmlPath, outputPath, *toDjango, *toGorm, *toEnt
}

func parseDbml(dbmlPath string) *core.DBML {
	file, err := os.Open(dbmlPath)
	if err != nil {
		panic(err)
	}

	scan := scanner.NewScanner(file)
	pars := parser.NewParser(scan)
	dbml, err := pars.Parse()
	if err != nil {
		panic(err)
	}
	return dbml
}

func main() {
	dbmlPath, outputPath, toDjango, toGorm, toEnt := parseArgs()

	dbml := parseDbml(dbmlPath)

	if toDjango {
		dbmldjango.CreateDjangoFiles(dbml, outputPath)
		fmt.Printf("Created Django models\nInput: %v\nOutput:%v\n", dbmlPath, outputPath)
	} else if toGorm {
		dbmlgorm.CreateGormFiles(dbml, outputPath)
		fmt.Printf("Created Gorm models\nInput: %v\nOutput:%v\n", dbmlPath, outputPath)
	} else if toEnt {
		dbmlent.CreateEntFiles(dbml, outputPath)
		fmt.Printf("Created Ent models\nInput: %v\nOutput:%v\n", dbmlPath, outputPath)
	}
}
