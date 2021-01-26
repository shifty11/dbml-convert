package common

import (
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
const PrefixCommon = "all:"
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

var commonRe = regexp.MustCompile(PrefixCommon + `\x60([^\x60]*)\x60`)
var djangoRe = regexp.MustCompile(PrefixDjango + `\x60([^\x60]*)\x60`)
var entRe = regexp.MustCompile(PrefixEnt + `\x60([^\x60]*)\x60`)

type SettingsType string

// State values.
const (
	DJangoSettings SettingsType = "DJangoSettings"
	EntSettings    SettingsType = "EntSettings"
)

func GetNoteSettings(note string, settingsType SettingsType) []string {
	var settings []string

	match := commonRe.FindStringSubmatch(note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], " ") {
			settings = append(settings, entry)
		}
	}
	var reg *regexp.Regexp
	if settingsType == DJangoSettings {
		reg = djangoRe
	} else {
		reg = entRe
	}
	match = reg.FindStringSubmatch(note)
	if len(match) == 2 {
		for _, entry := range strings.Split(match[1], " ") {
			settings = append(settings, entry)
		}
	}
	return settings
}
