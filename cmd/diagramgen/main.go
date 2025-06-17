package main

import (
	"diagramgen/pkg/parser"
	"diagramgen/pkg/renderer" // This package now contains Render (text) and RenderToPNG
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	// Read the example file
	content, err := ioutil.ReadFile("example.txt") // Assumes running from project root
	if err != nil {
		log.Fatalf("Error reading example.txt: %v", err)
	}

	// Parse the content
	allTablesData, err := parser.ParseAllText(string(content))
	if err != nil {
		log.Fatalf("Error parsing input: %v", err)
	}

	if allTablesData.MainTableID == "" {
		log.Fatalf("No main table ID found after parsing. Cannot determine which table to render.")
	}
	if len(allTablesData.Tables) == 0 {
		log.Fatalf("No tables parsed from input.")
	}

	mainTable, ok := allTablesData.Tables[allTablesData.MainTableID]
	if !ok {
		log.Fatalf("Main table with ID '%s' not found in parsed tables.", allTablesData.MainTableID)
	}

	// Render the main table to PNG
	outputFilePath := "output.png"
	err = renderer.RenderToPNG(&mainTable, allTablesData.Tables, outputFilePath) // Pass address of mainTable and all parsed tables
	if err != nil {
		log.Fatalf("Error rendering main table to PNG: %v", err)
	}

	fmt.Printf("Main table ('%s') successfully rendered to %s\n", allTablesData.MainTableID, outputFilePath)

	// Optional: Still render text version for comparison/debugging
	fmt.Println("\nTextual representation for debugging (main table):")
	textOutput := renderer.Render(mainTable) // Render takes table.Table by value
	fmt.Println(textOutput)
}
