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

// ExportMasterCookbook compiles all recipes into one giant PDF.
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

	// Cover Page
	pdf.AddPage()
	pdf.SetY(3.0)
	pdf.SetFont("Times", "B", 36)
	pdf.SetTextColor(107, 112, 92) // Sage Green
	pdf.CellFormat(0, 1.0, tr("Morris Family"), "", 1, "C", false, 0, "")
	pdf.SetFont("Times", "I", 24)
	pdf.CellFormat(0, 0.5, tr("Master Cookbook"), "", 1, "C", false, 0, "")

	for _, r := range recipes {
		pdf.AddPage()
		drawRecipePage(pdf, r, tr)
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
	pdf.CellFormat(0, 0.6, tr(r.Title), "", 1, "L", false, 0, "")

	// Tags
	pdf.SetFont("Times", "I", 10)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 0.3, tr(strings.Join(r.Tags, "  •  ")), "", 1, "L", false, 0, "")
	pdf.Ln(0.2)

	pdf.SetTextColor(0, 0, 0)

	// Ingredients Section
	pdf.SetFont("Times", "B", 14)
	pdf.CellFormat(0, 0.4, tr("INGREDIENTS"), "B", 1, "L", false, 0, "")
	pdf.Ln(0.1)

	pdf.SetFont("Times", "", 12)
	for _, ing := range r.Ingredients {
		if strings.TrimSpace(ing) != "" {
			pdf.CellFormat(0, 0.25, tr(" • "+ing), "", 1, "L", false, 0, "")
		}
	}
	pdf.Ln(0.3)

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
