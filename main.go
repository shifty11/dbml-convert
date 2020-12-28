package main

import (
	"flag"
	"fmt"
	"github.com/shifty111/dbml-to-gorm/dbmlgorm"
	"os"
)

func parseArgs() (string, string) {
	// Showing useful information when the user enters the --help option
	flag.Usage = func() {
		fmt.Printf("Usage: %s <path-to-dbml-file> <path-to-output-file>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	dbmlPath := flag.Arg(0)
	outputPath := flag.Arg(1)

	if dbmlPath == "" || outputPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return dbmlPath, outputPath
}

func main() {
	dbmlPath, outputPath := parseArgs()

	parsed := dbmlgorm.ParseDbml(dbmlPath)
	dbmlgorm.WriteToGormFile(parsed, outputPath)

	fmt.Printf("Created models from %v at %v\n", dbmlPath, outputPath)
}
