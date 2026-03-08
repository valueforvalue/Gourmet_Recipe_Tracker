package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ParseFile(path string) (Recipe, error) {
	file, err := os.Open(path)
	if err != nil {
		return Recipe{}, err
	}
	defer file.Close()

	recipe := Recipe{SourceFile: filepath.Base(path)}
	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		upperLine := strings.ToUpper(line)

		if strings.HasPrefix(upperLine, "RECIPE:") {
			recipe.Title = strings.TrimSpace(line[7:])
			continue
		} else if strings.HasPrefix(upperLine, "TAGS:") {
			tagString := strings.TrimSpace(line[5:])
			recipe.Tags = strings.Split(tagString, ",")
			continue
		} else if strings.Contains(upperLine, "INGREDIENTS") {
			currentSection = "INGREDIENTS"
			continue
		} else if strings.Contains(upperLine, "INSTRUCTIONS") {
			currentSection = "INSTRUCTIONS"
			continue
		} else if strings.HasPrefix(upperLine, "NOTES:") {
			recipe.Notes = strings.TrimSpace(line[6:])
			currentSection = "NOTES"
			continue
		}

		switch currentSection {
		case "INGREDIENTS":
			// Clean up leading dashes or bullet points
			item := strings.TrimPrefix(line, "- ")
			recipe.Ingredients = append(recipe.Ingredients, item)
		case "INSTRUCTIONS":
			// Strip existing numbers like "1." or "1. 1." before saving
			cleaned := cleanInstructionLine(line)
			recipe.Instructions = append(recipe.Instructions, cleaned)
		}
	}

	return recipe, scanner.Err()
}

// cleanInstructionLine uses a Regular Expression to remove leading numbers and periods
func cleanInstructionLine(line string) string {
	// This regex looks for digits followed by a period and a space at the start of a line
	// It will catch "1. ", "10. ", or even duplicated "1. 1. "
	re := regexp.MustCompile(`^(\d+\.\s*)+`)
	return strings.TrimSpace(re.ReplaceAllString(line, ""))
}
