package main

import (
	"flag"
	"fmt"
	"github.com/duythinht/dbml-go/core"
	"github.com/duythinht/dbml-go/parser"
	"github.com/duythinht/dbml-go/scanner"
	"github.com/shifty11/dbml-to-gorm/dbmldjango"
	"os"
)

func parseArgs() (string, string, string) {
	// Showing useful information when the user enters the --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s <path-to-dbml-file> <path-to-output-file> <path-to-django-models-file>\n",
			os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) < 3 {
		flag.Usage()
		os.Exit(0)
	}

	dbmlPath := flag.Arg(0)
	outputPath := flag.Arg(1)
	djangoPath := flag.Arg(2)

	if dbmlPath == "" || outputPath == "" || djangoPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return dbmlPath, outputPath, djangoPath
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
	dbmlPath, outputPath, _ := parseArgs()

	dbml := parseDbml(dbmlPath)

	//dbmlgorm.CreateGormFiles(dbml, outputPath)

	dbmldjango.CreateDjangoFiles(dbml, outputPath)

	fmt.Printf("Created models from %v at %v\n", dbmlPath, outputPath)
}
