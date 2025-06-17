package main

import (
	"diagramgen/pkg/parser"
	"diagramgen/pkg/renderer" // This package now contains Render (text) and RenderToPNG
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Define command-line flags
	inputFile := flag.String("inputFile", "example.txt", "Path to the input text file.")
	outputFile := flag.String("outputFile", "output.png", "Path to save the output PNG file.")
	verbose := flag.Bool("verbose", false, "Enable verbose logging.")

	// Shorthand flags
	flag.StringVar(inputFile, "i", "example.txt", "Path to the input text file (shorthand).")
	flag.StringVar(outputFile, "o", "output.png", "Path to save the output PNG file (shorthand).")
	flag.BoolVar(verbose, "v", false, "Enable verbose logging (shorthand).")

	flag.Parse()

	if *verbose {
		log.Printf("Input file: %s", *inputFile)
		log.Printf("Output file: %s", *outputFile)
		log.Printf("Verbose logging enabled")
	}

	// Read the input file
	content, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		log.Printf("Error reading input file '%s': %v", *inputFile, err)
		os.Exit(1)
	}

	if *verbose {
		log.Printf("Successfully read input file: %s", *inputFile)
	}

	// Parse the content
	allTablesData, err := parser.ParseAllText(string(content))
	if err != nil {
		log.Printf("Error parsing input from file '%s': %v", *inputFile, err)
		os.Exit(1)
	}

	if *verbose {
		log.Printf("Successfully parsed content from file: %s", *inputFile)
	}

	if allTablesData.MainTableID == "" {
		log.Println("Error: No main table ID found after parsing. Cannot determine which table to render.")
		os.Exit(1)
	}
	if len(allTablesData.Tables) == 0 {
		log.Println("Error: No tables parsed from input.")
		os.Exit(1)
	}

	mainTable, ok := allTablesData.Tables[allTablesData.MainTableID]
	if !ok {
		log.Printf("Error: Main table with ID '%s' not found in parsed tables.", allTablesData.MainTableID)
		os.Exit(1)
	}

	if *verbose {
		log.Printf("Main table to render: %s", allTablesData.MainTableID)
	}

	// Render the main table to PNG
	err = renderer.RenderToPNG(&mainTable, allTablesData.Tables, *outputFile) // Pass address of mainTable and all parsed tables
	if err != nil {
		log.Printf("Error rendering main table to PNG '%s': %v", *outputFile, err)
		os.Exit(1)
	}

	fmt.Printf("Main table ('%s') successfully rendered to %s\n", allTablesData.MainTableID, *outputFile)

	// Optional: Still render text version for comparison/debugging if verbose
	if *verbose {
		log.Println("\nTextual representation for debugging (main table):")
		textOutput := renderer.Render(mainTable) // Render takes table.Table by value
		fmt.Println(textOutput)                  // Use fmt.Println to send to stdout directly, respecting verbose for logs
	}
}
