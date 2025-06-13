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
	tableData, err := parser.Parse(string(content))
	if err != nil {
		log.Fatalf("Error parsing table: %v", err)
	}

	// Render the table to PNG
	outputFilePath := "output.png"
	err = renderer.RenderToPNG(&tableData, outputFilePath) // Pass address of tableData
	if err != nil {
		// Log the error but don't necessarily make it fatal if text rendering can still proceed.
		// However, for this task, if PNG rendering fails, it's a primary concern.
		log.Fatalf("Error rendering table to PNG: %v", err)
	}

	fmt.Printf("Table successfully rendered to %s\n", outputFilePath)

	// Optional: Still render text version for comparison/debugging
	// This helps verify parsing independently of PNG rendering issues (e.g. font missing)
	fmt.Println("\nTextual representation for debugging:")
	textOutput := renderer.Render(tableData) // Text renderer from the same package
	fmt.Println(textOutput)
}
