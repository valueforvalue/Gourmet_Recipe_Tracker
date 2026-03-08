package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func ExportToPDF(r Recipe, isBooklet bool) error {
	outputDir := "Printables"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	var pdf *gofpdf.Fpdf
	var sidebarWidth, contentWidth, startX float64
	var maxTitleWidth float64

	// 1. Setup Page Dimensions (Both now use Inches for North American Standards)
	if isBooklet {
		// Half-Letter Size: 5.5 x 8.5 inches
		pdf = gofpdf.NewCustom(&gofpdf.InitType{
			UnitStr: "in",
			Size:    gofpdf.SizeType{Wd: 5.5, Ht: 8.5},
		})
		pdf.SetMargins(0.5, 0.5, 0.5)
		sidebarWidth = 1.5
		contentWidth = 2.8
		startX = 0.6
		maxTitleWidth = 4.4
	} else {
		// Standard Letter Size: 8.5 x 11 inches
		pdf = gofpdf.New("P", "in", "Letter", "")
		pdf.SetMargins(0.75, 0.75, 0.75)
		sidebarWidth = 2.2
		contentWidth = 4.5
		startX = 1.0
		maxTitleWidth = 7.0
	}

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// 2. Set Footer (defined before AddPage)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-0.5) // Half inch from bottom
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 0.4, tr(fmt.Sprintf("Page %d", pdf.PageNo())), "", 0, "C", false, 0, "")
	})

	// 3. Set Auto Page Break (Both use inches now)
	pdf.SetAutoPageBreak(true, 0.75)

	pdf.AddPage()

	// 4. Define Visual Constants (Inches-friendly scaling)
	titleSize := 32.0
	bodySize := 11.0
	headerSize := 13.0
	accentWidth := 0.4
	rowHeight := 0.5
	tagHeight := 0.25

	if isBooklet {
		titleSize = 20.0
		bodySize = 9.0
		headerSize = 11.0
		accentWidth = 0.2
		rowHeight = 0.35
		tagHeight = 0.2
	}

	// --- DECORATIVE ACCENT ---
	pdf.SetFillColor(60, 60, 60)
	pdf.Rect(0, 0, accentWidth, 11.0, "F")

	// --- HEADER (TITLE & TAGS) ---
	pdf.SetX(startX)
	pdf.SetFont("Times", "B", titleSize)
	pdf.SetTextColor(40, 40, 40)

	// MultiCell allows the title to wrap
	pdf.MultiCell(maxTitleWidth, rowHeight, tr(r.Title), "", "L", false)

	pdf.SetX(startX)
	pdf.SetFont("Arial", "I", bodySize-1)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, tagHeight, tr(strings.Join(r.Tags, "  •  ")), "", 1, "L", false, 0, "")

	pdf.Ln(0.2)

	columnStartY := pdf.GetY()

	// --- SIDEBAR: INGREDIENTS ---
	pdf.SetFont("Arial", "B", headerSize)
	pdf.SetTextColor(80, 20, 20)
	pdf.SetX(startX)

	headerH := 0.35
	pdf.Cell(sidebarWidth, headerH, tr("INGREDIENTS"))
	pdf.Ln(headerH)

	pdf.SetFont("Arial", "", bodySize)
	pdf.SetTextColor(0, 0, 0)

	ingH := 0.22

	for _, ing := range r.Ingredients {
		pdf.SetX(startX)
		pdf.MultiCell(sidebarWidth, ingH, tr("• "+ing), "", "L", false)
		pdf.Ln(0.05)
	}

	// --- MAIN CONTENT: PREPARATION ---
	pdf.SetY(columnStartY)

	gutter := 0.3
	mainColX := startX + sidebarWidth + gutter

	pdf.SetX(mainColX)
	pdf.SetFont("Arial", "B", headerSize)
	pdf.SetTextColor(80, 20, 20)
	pdf.Cell(0, headerH, tr("PREPARATION"))
	pdf.Ln(headerH)

	stepH := 0.25

	for i, step := range r.Instructions {
		pdf.SetX(mainColX)
		pdf.SetFont("Times", "B", bodySize+1)
		numStr := fmt.Sprintf("%d. ", i+1)

		numW := 0.25
		pdf.Cell(numW, stepH, numStr)

		pdf.SetFont("Times", "", bodySize+1)
		pdf.MultiCell(contentWidth, stepH, tr(step), "", "L", false)
		pdf.Ln(0.12)
	}

	// --- NOTES ---
	if r.Notes != "" {
		pdf.Ln(0.2)
		pdf.SetX(mainColX)
		pdf.SetFillColor(255, 255, 240)
		pdf.SetFont("Arial", "I", bodySize)
		pdf.MultiCell(contentWidth, ingH, tr("Cook's Note: "+r.Notes), "L", "L", true)
	}

	// 5. Output File
	suffix := ""
	if isBooklet {
		suffix = "_Booklet"
	}
	outputPath := filepath.Join(outputDir, r.Title+suffix+".pdf")
	return pdf.OutputFileAndClose(outputPath)
}
