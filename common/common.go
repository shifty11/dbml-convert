package common

import (
	"github.com/duythinht/dbml-go/core"
	"os"
	"regexp"
	"strings"
)

// types in dbml
const TDecimal = "decimal"
const TString = "string"
const TUint = "uint"
const TInt = "int"
const TEmail = "email"
const TDatetime = "datetime"

// settings in dbml notes
const SHidden = "hidden"
const SBackref = "backref"

// Prefixes in dbml notes
const PrefixDjango = "django:"
const PrefixEnt = "ent:"

func WriteToFile(data string, outputPath string) {
	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(data)
	if err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}

//const matchChars = "[a-zA-Z0-9-_;:<>= ./'\"%&!?]"
const matchChars = "[a-zA-Z0-9-_;:<>= ./'\"%&!?]"

var commonRe = regexp.MustCompile(`^([a-zA-Z0-9-_;:<>= ./'"%&!?]*)$`)

//var djangoRe = regexp.MustCompile(`^` + PrefixDjango + `(` + matchChars + `*)$`)
//var entRe = regexp.MustCompile(`^` + PrefixEnt + `(` + matchChars + `*)$`)

func GetColumnSettings(column core.Column) []string {
	var settings []string
	//if strings.HasPrefix(column.Settings.Note, PrefixDjango) || strings.HasPrefix(column.Settings.Note, PrefixEnt) {
	//	return settings
	//}

	match := commonRe.FindStringSubmatch(column.Settings.Note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], " ") {
			settings = append(settings, strings.ToLower(entry))
		}
	}
	return settings
}

//func GetSpecialDjangoParameters(column core.Column) []string {
//	var parameters []string
//	if !strings.HasPrefix(column.Settings.Note, PrefixDjango) {
//		return parameters
//	}
//	match := commonRe.FindStringSubmatch(column.Settings.Note[len(PrefixDjango):])
//	if len(match) == 2 {
//		for _, entry := range strings.Split(match[1], " ") {
//			parameters = append(parameters, strings.ToLower(entry))
//		}
//	}
//	return parameters
//}
//
//func GetSpecialEntParameters(column core.Column) []string {
//	var parameters []string
//	if !strings.HasPrefix(column.Settings.Note, PrefixEnt) {
//		return parameters
//	}
//	match := commonRe.FindStringSubmatch(column.Settings.Note[len(PrefixEnt):])
//	if len(match) == 2 {
//		for _, entry := range strings.Split(match[1], " ") {
//			parameters = append(parameters, strings.ToLower(entry))
//		}
//	}
//	return parameters
//}
