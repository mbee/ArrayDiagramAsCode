package main

import (
	"diagramgen/pkg/parser"
	"diagramgen/pkg/renderer"
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

	// Render the table
	output := renderer.Render(tableData)

	// Print the output
	fmt.Println(output)
}
