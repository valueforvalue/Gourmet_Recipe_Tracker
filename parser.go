// File: parser.go
package main

import (
	"bufio"
	"os"
	"strings"
)

// ParseRecipeFile reads a standard Gourmet Tracker text file and converts it into a Recipe struct.
func ParseRecipeFile(filename string) (Recipe, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Recipe{}, err
	}
	defer file.Close()

	var r Recipe
	// The legacy SourceFile field has been removed to match models.go

	scanner := bufio.NewScanner(file)
	var currentSection string

	for scanner.Scan() {
		line := scanner.Text()
		lineTrim := strings.TrimSpace(line)

		if lineTrim == "" {
			continue
		}

		// Detect Section Headers
		if strings.HasPrefix(line, "RECIPE:") {
			r.Title = strings.TrimSpace(strings.TrimPrefix(line, "RECIPE:"))
			continue
		}
		if strings.HasPrefix(line, "TAGS:") {
			tagList := strings.TrimPrefix(line, "TAGS:")
			parts := strings.Split(tagList, ",")
			for _, p := range parts {
				cleanTag := strings.TrimSpace(p)
				if cleanTag != "" { // Prevents empty tags from trailing commas
					r.Tags = append(r.Tags, cleanTag)
				}
			}
			continue
		}

		// Section Switching
		if lineTrim == "INGREDIENTS" {
			currentSection = "ingredients"
			continue
		}
		if lineTrim == "INSTRUCTIONS" {
			currentSection = "instructions"
			continue
		}
		if strings.HasPrefix(line, "NOTES:") {
			r.Notes = strings.TrimSpace(strings.TrimPrefix(line, "NOTES:"))
			currentSection = ""
			continue
		}

		// Append data based on current section
		switch currentSection {
		case "ingredients":
			// Strip leading dashes or bullets if present
			item := strings.TrimPrefix(lineTrim, "-")
			r.Ingredients = append(r.Ingredients, strings.TrimSpace(item))
		case "instructions":
			r.Instructions = append(r.Instructions, lineTrim)
		}
	}

	if err := scanner.Err(); err != nil {
		return Recipe{}, err
	}

	return r, nil
}
