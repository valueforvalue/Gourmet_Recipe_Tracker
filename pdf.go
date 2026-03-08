package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// ExportMasterCookbook generates one PDF containing all recipes with a Table of Contents
func ExportMasterCookbook(recipes []Recipe) error {
	outputDir := "Printables"
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "in",
		Size:    gofpdf.SizeType{Wd: 5.5, Ht: 8.5},
	})

	// Set 0.75" Left margin for binding/punching holes
	pdf.SetMargins(0.75, 0.5, 0.5)
	pdf.SetAutoPageBreak(true, 0.75)
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// 1. Footer with Page Numbers
	pdf.SetFooterFunc(func() {
		pdf.SetY(-0.5)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 0.4, tr(fmt.Sprintf("Page %d", pdf.PageNo())), "", 0, "C", false, 0, "")
	})

	// 2. Table of Contents
	pdf.AddPage()
	pdf.SetFont("Times", "B", 24)
	pdf.CellFormat(0, 1, tr("Our Family Cookbook"), "", 1, "C", false, 0, "")
	pdf.Ln(0.5)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 0.4, tr("Table of Contents"))
	pdf.Ln(0.4)
	pdf.SetFont("Arial", "", 10)

	// Create the TOC entries
	for _, r := range recipes {
		pdf.CellFormat(0, 0.25, tr(r.Title), "B", 1, "L", false, 0, "")
	}

	// 3. Add each recipe
	for _, r := range recipes {
		addRecipeToPDF(pdf, r, true, tr)
	}

	return pdf.OutputFileAndClose(filepath.Join(outputDir, "Master_Cookbook.pdf"))
}

// ExportToPDF remains for single file exports
func ExportToPDF(r Recipe, isBooklet bool) error {
	outputDir := "Printables"
	var pdf *gofpdf.Fpdf
	if isBooklet {
		pdf = gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "in", Size: gofpdf.SizeType{Wd: 5.5, Ht: 8.5}})
		pdf.SetMargins(0.75, 0.5, 0.5)
		pdf.SetAutoPageBreak(true, 0.75)
	} else {
		pdf = gofpdf.New("P", "in", "Letter", "")
		pdf.SetMargins(0.75, 0.75, 0.75)
		pdf.SetAutoPageBreak(true, 0.75)
	}

	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-0.5)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 0.4, tr(fmt.Sprintf("Page %d", pdf.PageNo())), "", 0, "C", false, 0, "")
	})

	addRecipeToPDF(pdf, r, isBooklet, tr)

	suffix := ""
	if isBooklet {
		suffix = "_Booklet"
	}
	return pdf.OutputFileAndClose(filepath.Join(outputDir, r.Title+suffix+".pdf"))
}

// Shared logic for drawing the actual recipe content
func addRecipeToPDF(pdf *gofpdf.Fpdf, r Recipe, isBooklet bool, tr func(string) string) {
	pdf.AddPage()

	// Constants - Initialized with Standard Letter defaults
	startX := 1.0
	titleSize := 32.0
	bodySize := 11.0
	headerSize := 13.0
	accentWidth := 0.4
	rowH := 0.5
	sidebarW := 2.2
	contentW := 4.5
	gutter := 0.3

	// Override if Booklet mode
	if isBooklet {
		startX = 0.75
		titleSize = 20.0
		bodySize = 9.0
		headerSize = 11.0
		accentWidth = 0.2
		rowH = 0.35
		sidebarW = 1.4
		contentW = 2.6
	}

	// Accent Bar
	pdf.SetFillColor(60, 60, 60)
	pdf.Rect(0, 0, accentWidth, 11, "F")

	// Header
	pdf.SetX(startX)
	pdf.SetFont("Times", "B", titleSize)
	pdf.SetTextColor(40, 40, 40)
	pdf.MultiCell(0, rowH, tr(r.Title), "", "L", false)

	pdf.SetX(startX)
	pdf.SetFont("Arial", "I", bodySize-1)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 0.2, tr(strings.Join(r.Tags, "  •  ")), "", 1, "L", false, 0, "")
	pdf.Ln(0.2)

	colY := pdf.GetY()

	// Ingredients
	pdf.SetX(startX)
	pdf.SetFont("Arial", "B", headerSize)
	pdf.SetTextColor(80, 20, 20)
	pdf.Cell(sidebarW, 0.35, tr("INGREDIENTS"))
	pdf.Ln(0.35)
	pdf.SetFont("Arial", "", bodySize)
	pdf.SetTextColor(0, 0, 0)
	for _, ing := range r.Ingredients {
		pdf.SetX(startX)
		pdf.MultiCell(sidebarW, 0.2, tr("• "+ing), "", "L", false)
		pdf.Ln(0.05)
	}

	// Preparation
	pdf.SetY(colY)
	pdf.SetX(startX + sidebarW + gutter)
	pdf.SetFont("Arial", "B", headerSize)
	pdf.SetTextColor(80, 20, 20)
	pdf.Cell(contentW, 0.35, tr("PREPARATION"))
	pdf.Ln(0.35)
	pdf.SetFont("Times", "", bodySize+1)
	pdf.SetTextColor(30, 30, 30)
	for i, step := range r.Instructions {
		pdf.SetX(startX + sidebarW + gutter)
		pdf.SetFont("Times", "B", bodySize+1)
		pdf.Cell(0.25, 0.25, fmt.Sprintf("%d. ", i+1))
		pdf.SetFont("Times", "", bodySize+1)
		pdf.MultiCell(contentW, 0.25, tr(step), "", "L", false)
		pdf.Ln(0.12)
	}

	// Notes
	if r.Notes != "" {
		pdf.Ln(0.2)
		pdf.SetX(startX + sidebarW + gutter)
		pdf.SetFillColor(255, 255, 240)
		pdf.SetFont("Arial", "I", bodySize)
		pdf.MultiCell(contentW, 0.2, tr("Cook's Note: "+r.Notes), "L", "L", true)
	}
}
