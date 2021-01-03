package common

import "os"

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
