// File: pdf.go
package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// ExportToPDF generates a single recipe PDF in either Letter or Booklet format.
func ExportToPDF(r Recipe, isBooklet bool) error {
	outputDir := "Printables"
	var pdf *gofpdf.Fpdf

	if isBooklet {
		pdf = gofpdf.NewCustom(&gofpdf.InitType{
			UnitStr: "in",
			Size:    gofpdf.SizeType{Wd: 5.5, Ht: 8.5},
		})
		pdf.SetMargins(0.5, 0.5, 0.5)
	} else {
		pdf = gofpdf.New("P", "in", "Letter", "")
		pdf.SetMargins(0.75, 0.75, 0.75)
	}

	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetAutoPageBreak(true, 0.75)

	// Footer with Page Numbers
	pdf.SetFooterFunc(func() {
		pdf.SetY(-0.5)
		pdf.SetFont("Times", "I", 8)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(0, 0.4, tr(fmt.Sprintf("Morris Family Recipe Box - Page %d", pdf.PageNo())), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	drawRecipePage(pdf, r, tr)

	suffix := "_Letter.pdf"
	if isBooklet {
		suffix = "_Booklet.pdf"
	}

	outputPath := filepath.Join(outputDir, r.Title+suffix)
	return pdf.OutputFileAndClose(outputPath)
}

// ExportMasterCookbook compiles all recipes into one giant PDF, grouped by tags.
func ExportMasterCookbook(recipes []Recipe, isBooklet bool) error {
	outputDir := "Printables"
	var pdf *gofpdf.Fpdf

	if isBooklet {
		pdf = gofpdf.NewCustom(&gofpdf.InitType{
			UnitStr: "in",
			Size:    gofpdf.SizeType{Wd: 5.5, Ht: 8.5},
		})
		pdf.SetMargins(0.5, 0.5, 0.5)
	} else {
		pdf = gofpdf.New("P", "in", "Letter", "")
		pdf.SetMargins(0.75, 0.75, 0.75)
	}

	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetAutoPageBreak(true, 0.75)

	// --- 1. Cover Page ---
	pdf.AddPage()
	pdf.SetY(3.0)
	pdf.SetFont("Times", "B", 36)
	pdf.SetTextColor(107, 112, 92) // Sage Green
	pdf.CellFormat(0, 1.0, tr("Morris Family"), "", 1, "C", false, 0, "")
	pdf.SetFont("Times", "I", 24)
	pdf.CellFormat(0, 0.5, tr("Master Cookbook"), "", 1, "C", false, 0, "")

	// --- 2. Group recipes by tags ---
	groupedRecipes := make(map[string][]Recipe)
	for _, r := range recipes {
		if len(r.Tags) == 0 {
			groupedRecipes["Miscellaneous"] = append(groupedRecipes["Miscellaneous"], r)
		} else {
			for _, tag := range r.Tags {
				cleanTag := strings.TrimSpace(tag)
				if cleanTag != "" {
					groupedRecipes[cleanTag] = append(groupedRecipes[cleanTag], r)
				}
			}
		}
	}

	// --- 3. Get sorted list of chapters ---
	var chapters []string
	hasMisc := false
	for tag := range groupedRecipes {
		if tag == "Miscellaneous" {
			hasMisc = true
			continue
		}
		chapters = append(chapters, tag)
	}
	sort.Strings(chapters)
	if hasMisc {
		chapters = append(chapters, "Miscellaneous")
	}

	// --- 4. Table of Contents Page ---
	pdf.AddPage()
	pdf.SetY(1.0)
	pdf.SetFont("Times", "B", 24)
	pdf.SetTextColor(107, 112, 92)
	pdf.CellFormat(0, 0.8, tr("Table of Contents"), "B", 1, "C", false, 0, "")
	pdf.Ln(0.5)

	pdf.SetFont("Times", "", 14)
	pdf.SetTextColor(0, 0, 0)
	for _, chapter := range chapters {
		count := len(groupedRecipes[chapter])
		pdf.CellFormat(0, 0.35, tr(fmt.Sprintf("%s (%d recipes)", chapter, count)), "", 1, "L", false, 0, "")
		pdf.Ln(0.05)
	}

	// --- 5. Generate Chapter Pages and Recipes ---
	for _, chapter := range chapters {
		// Chapter Divider Page
		pdf.AddPage()
		if isBooklet {
			pdf.SetY(3.5)
		} else {
			pdf.SetY(4.5)
		}
		pdf.SetFont("Times", "B", 28)
		pdf.SetTextColor(107, 112, 92) // Sage Green
		pdf.CellFormat(0, 1.0, tr(chapter), "", 1, "C", false, 0, "")

		// Recipe Pages for this chapter
		pdf.SetTextColor(0, 0, 0) // Reset to black
		for _, r := range groupedRecipes[chapter] {
			pdf.AddPage()
			drawRecipePage(pdf, r, tr)
		}
	}

	fileName := "Master_Cookbook_Full.pdf"
	if isBooklet {
		fileName = "Master_Cookbook_Booklet.pdf"
	}

	return pdf.OutputFileAndClose(filepath.Join(outputDir, fileName))
}

// drawRecipePage handles the visual layout logic.
func drawRecipePage(pdf *gofpdf.Fpdf, r Recipe, tr func(string) string) {
	// Title - Sage Green (#6B705C)
	pdf.SetFont("Times", "B", 22)
	pdf.SetTextColor(107, 112, 92)
	pdf.MultiCell(0, 0.35, tr(r.Title), "", "L", false)
	pdf.Ln(0.1)

	// Tags
	pdf.SetFont("Times", "I", 10)
	pdf.SetTextColor(128, 128, 128)
	pdf.MultiCell(0, 0.15, tr(strings.Join(r.Tags, "  •  ")), "", "L", false)
	pdf.Ln(0.25)

	pdf.SetTextColor(0, 0, 0)

	// Ingredients Section
	pdf.SetFont("Times", "B", 14)
	pdf.CellFormat(0, 0.4, tr("INGREDIENTS"), "B", 1, "L", false, 0, "")
	pdf.Ln(0.1)

	pdf.SetFont("Times", "", 12)
	for _, ing := range r.Ingredients {
		if strings.TrimSpace(ing) != "" {
			pdf.MultiCell(0, 0.2, tr(" • "+ing), "", "L", false)
			pdf.Ln(0.05)
		}
	}
	pdf.Ln(0.2)

	// Preparation Section
	pdf.SetFont("Times", "B", 14)
	pdf.CellFormat(0, 0.4, tr("PREPARATION"), "B", 1, "L", false, 0, "")
	pdf.Ln(0.1)

	pdf.SetFont("Times", "", 12)
	re := regexp.MustCompile(`^\d+[\.\)]\s*`)

	for i, step := range r.Instructions {
		if strings.TrimSpace(step) != "" {
			cleanStep := re.ReplaceAllString(strings.TrimSpace(step), "")
			pdf.MultiCell(0, 0.25, tr(fmt.Sprintf("%d. %s", i+1, cleanStep)), "", "L", false)
			pdf.Ln(0.1)
		}
	}

	// Notes Section
	if r.Notes != "" {
		pdf.Ln(0.2)
		pdf.SetFont("Times", "I", 10)
		pdf.SetTextColor(80, 80, 80)
		pdf.MultiCell(0, 0.2, tr("Notes: "+r.Notes), "T", "L", false)
	}
}
