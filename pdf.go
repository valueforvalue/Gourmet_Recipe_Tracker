package main

import (
	"fmt"
	"path/filepath"
	"regexp"
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

	pdf.SetFooterFunc(func() {
		pdf.SetY(-0.5)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 0.4, tr(fmt.Sprintf("Page %d", pdf.PageNo())), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	drawRecipePage(pdf, r, isBooklet, tr)

	suffix := "_Letter.pdf"
	if isBooklet {
		suffix = "_Booklet.pdf"
	}

	outputPath := filepath.Join(outputDir, r.Title+suffix)
	return pdf.OutputFileAndClose(outputPath)
}

// ExportMasterCookbook loops through all recipes and compiles them into one giant PDF.
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

	for _, r := range recipes {
		pdf.AddPage()
		drawRecipePage(pdf, r, isBooklet, tr)
	}

	fileName := "Master_Cookbook_Full.pdf"
	if isBooklet {
		fileName = "Master_Cookbook_Booklet.pdf"
	}

	return pdf.OutputFileAndClose(filepath.Join(outputDir, fileName))
}

// drawRecipePage is the internal layout engine for the PDF content.
func drawRecipePage(pdf *gofpdf.Fpdf, r Recipe, isBooklet bool, tr func(string) string) {
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 0.4, tr(r.Title), "", 1, "L", false, 0, "")

	// Tags
	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 0.3, tr(fmt.Sprintf("Tags: %v", strings.Join(r.Tags, ", "))), "", 1, "L", false, 0, "")
	pdf.Ln(0.2)

	pdf.SetTextColor(0, 0, 0)

	// Ingredients Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 0.3, tr("Ingredients"))
	pdf.Ln(0.3)
	pdf.SetFont("Arial", "", 11)
	for _, ing := range r.Ingredients {
		pdf.CellFormat(0, 0.2, tr("- "+ing), "", 1, "L", false, 0, "")
	}
	pdf.Ln(0.3)

	// Instructions Section (With Number Scrubbing)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 0.3, tr("Instructions"))
	pdf.Ln(0.3)
	pdf.SetFont("Arial", "", 11)

	// This Regex looks for leading digits followed by a period or closing parenthesis and a space
	re := regexp.MustCompile(`^\d+[\.\)]\s*`)

	for i, step := range r.Instructions {
		// Clean the step by removing existing numbers if they exist
		cleanStep := re.ReplaceAllString(strings.TrimSpace(step), "")

		pdf.MultiCell(0, 0.2, tr(fmt.Sprintf("%d. %s", i+1, cleanStep)), "", "L", false)
		pdf.Ln(0.1)
	}

	// Notes Section
	if r.Notes != "" {
		pdf.Ln(0.2)
		pdf.SetFont("Arial", "I", 10)
		pdf.MultiCell(0, 0.2, tr("Notes: "+r.Notes), "1", "L", false)
	}
}
