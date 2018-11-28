package main

import (
	"flag"
	"log"
	"os"
	"path"
	"strings"

	"github.com/stephen-fox/vmwareify"
)

const (
	inputFilePathArg  = "f"
	outputFilePathArg = "o"
	helpArg           = "h"
)

func main() {
	inputFilePath := flag.String(inputFilePathArg, "", "The .ovf file to convert")
	outputFilePath := flag.String(outputFilePathArg, "", "The output file path for the converted file")
	help := flag.Bool(helpArg, false, "Display this help page")

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if len(*inputFilePath) == 0 {
		log.Fatal("Please specify a .ovf file to convert")
	}

	if len(*outputFilePath) == 0 {
		inputFilename := path.Base(*inputFilePath)
		*outputFilePath = path.Dir(*inputFilePath) + "/" + getFilenameWithoutExtension(inputFilename) + "-vmware" + getFileExtension(inputFilename)
	}

	err := vmwareify.BasicConvert(*inputFilePath, *outputFilePath)
	if err != nil {
		log.Fatal("Failed to convert .ovf file - " + err.Error())
	}

	log.Println("Saved converted file to '" + *outputFilePath + "'")
}

func getFilenameWithoutExtension(filename string) string {
	index := strings.LastIndex(filename, ".")

	if index > 0 {
		return filename[:index]
	}

	return ""
}

func getFileExtension(filename string) string {
	index := strings.LastIndex(filename, ".")

	if index > 0 {
		return filename[index:]
	}

	return ".ovf"
}
