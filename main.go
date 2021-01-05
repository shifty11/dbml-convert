package main

import (
	"flag"
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/duythinht/dbml-go/parser"
	"github.com/duythinht/dbml-go/scanner"
	"github.com/shifty11/dbml-to-gorm/dbmldjango"
	"github.com/shifty11/dbml-to-gorm/dbmlgorm"
	"os"
)

func parseArgs() (string, string, bool) {
	// Showing useful information when the user enters the --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s <path-to-dbml-file> <path-to-output>\n",
			os.Args[0])
		flag.PrintDefaults()
	}
	toDjango := flag.Bool("django", false, "Creates Django models")
	toGorm := flag.Bool("gorm", false, "Creates Gorm models")
	flag.Parse()

	if len(flag.Args()) < 2 || (!*toDjango && !*toGorm) || (*toDjango && *toGorm) {
		flag.Usage()
		os.Exit(1)
	}

	dbmlPath := flag.Arg(0)
	outputPath := flag.Arg(1)

	if dbmlPath == "" || outputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return dbmlPath, outputPath, *toDjango
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
	dbmlPath, outputPath, toDjango := parseArgs()

	dbml := parseDbml(dbmlPath)

	if toDjango {
		dbmldjango.CreateDjangoFiles(dbml, outputPath)
		fmt.Printf("Created Django models from %v at %v\n", dbmlPath, outputPath)
	} else {
		dbmlgorm.CreateGormFiles(dbml, outputPath)
		fmt.Printf("Created Gorm models from %v at %v\n", dbmlPath, outputPath)
	}
}
