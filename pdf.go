package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func ExportToPDF(r Recipe) error {
	outputDir := "Printables"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.SetMargins(15, 20, 15)
	pdf.AddPage()

	// --- DESIGN ACCENTS ---
	// Add a decorative border/line on the far left for flair
	pdf.SetFillColor(60, 60, 60)
	pdf.Rect(0, 0, 10, 297, "F")

	// --- HEADER ---
	pdf.SetX(20)
	pdf.SetFont("Times", "B", 32)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(0, 20, tr(r.Title), "", 1, "L", false, 0, "")

	// Tags as a clean sub-header
	pdf.SetX(20)
	pdf.SetFont("Arial", "I", 11)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 5, tr(strings.Join(r.Tags, "  •  ")), "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// Save Y position for the two-column start
	columnStartY := pdf.GetY()

	// --- COLUMN 1: INGREDIENTS (The Sidebar) ---
	pdf.SetY(columnStartY)
	pdf.SetX(20)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(80, 20, 20) // Deep dark red for headers
	pdf.Cell(50, 10, tr("INGREDIENTS"))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(0, 0, 0)
	for _, ing := range r.Ingredients {
		pdf.SetX(20)
		// MultiCell allows long ingredients to wrap within the sidebar width
		pdf.MultiCell(50, 5, tr("• "+ing), "", "L", false)
		pdf.Ln(1)
	}

	// Draw a vertical divider line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(75, columnStartY, 75, 270)

	// --- COLUMN 2: PREPARATION (The Main Body) ---
	pdf.SetY(columnStartY)
	pdf.SetX(80) // Shift to the right of the divider

	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(80, 20, 20)
	pdf.Cell(0, 10, tr("PREPARATION"))
	pdf.Ln(8)

	pdf.SetFont("Times", "", 12) // Classic serif font for the body text
	pdf.SetTextColor(30, 30, 30)
	for i, step := range r.Instructions {
		pdf.SetX(80)
		pdf.SetFont("Times", "B", 12)
		pdf.Cell(8, 6, fmt.Sprintf("%d.", i+1))

		pdf.SetFont("Times", "", 12)
		// Use a wider width for the main column
		pdf.MultiCell(110, 6, tr(step), "", "L", false)
		pdf.Ln(4)
	}

	// --- FOOTER: NOTES ---
	if r.Notes != "" {
		pdf.SetY(-40) // Place near the bottom
		pdf.SetX(80)
		pdf.SetFillColor(245, 245, 240)
		pdf.SetDrawColor(80, 20, 20)
		pdf.SetLineWidth(0.5)
		pdf.SetFont("Arial", "I", 10)
		pdf.MultiCell(110, 6, tr("Cook's Note: "+r.Notes), "L", "L", true)
	}

	outputPath := filepath.Join(outputDir, r.Title+".pdf")
	return pdf.OutputFileAndClose(outputPath)
}
