package main

import (
	"flag"
	"log"
	"os"

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
		log.Fatal("Please specify the full output file path")
	}

	err := vmwareify.BasicConvert(*inputFilePath, *outputFilePath)
	if err != nil {
		log.Fatal("Failed to convert .ovf file - " + err.Error())
	}
}
