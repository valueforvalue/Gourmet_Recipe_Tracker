package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func ExportMasterCookbook(recipes []Recipe, isBooklet bool) error {
	outputDir := "Printables"
	var pdf *gofpdf.Fpdf

	if isBooklet {
		pdf = gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "in", Size: gofpdf.SizeType{Wd: 5.5, Ht: 8.5}})
		pdf.SetMargins(0.75, 0.5, 0.5)
	} else {
		pdf = gofpdf.New("P", "in", "Letter", "")
		pdf.SetMargins(0.75, 0.75, 0.75)
	}

	pdf.SetAutoPageBreak(true, 0.75)
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.SetFooterFunc(func() {
		pdf.SetY(-0.5)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 0.4, tr(fmt.Sprintf("Page %d", pdf.PageNo())), "", 0, "C", false, 0, "")
	})

	// 1. Create Internal Links
	links := make(map[string]int)
	for _, r := range recipes {
		links[r.Title] = pdf.AddLink()
	}

	// 2. Table of Contents
	pdf.AddPage()
	titleFS, tocHeaderFS, tocItemFS := 30.0, 14.0, 12.0
	if isBooklet {
		titleFS, tocHeaderFS, tocItemFS = 20.0, 11.0, 9.0
	}

	pdf.SetFont("Times", "B", titleFS)
	pdf.CellFormat(0, 1.0, tr("Family Recipe Collection"), "", 1, "C", false, 0, "")
	pdf.Ln(0.2)
	pdf.SetFont("Arial", "B", tocHeaderFS)
	pdf.SetTextColor(80, 20, 20)
	pdf.Cell(0, 0.4, tr("Table of Contents"))
	pdf.Ln(0.3)

	pdf.SetFont("Arial", "", tocItemFS)
	pdf.SetTextColor(0, 0, 255)
	for _, r := range recipes {
		pdf.WriteLinkID(0.3, tr("• "+r.Title), links[r.Title])
		if isBooklet {
			pdf.Ln(0.25)
		} else {
			pdf.Ln(0.3)
		}
	}
	pdf.SetTextColor(0, 0, 0)

	// 3. Add Recipe Pages - FIX: Ensure AddPage is inside the loop
	for _, r := range recipes {
		pdf.AddPage() // Forces each recipe to start on a fresh sheet
		pdf.SetLink(links[r.Title], 0, -1)
		drawRecipePage(pdf, r, isBooklet, tr)
	}

	fileName := "Master_Cookbook_Full.pdf"
	if isBooklet {
		fileName = "Master_Cookbook_Booklet.pdf"
	}
	return pdf.OutputFileAndClose(filepath.Join(outputDir, fileName))
}

func ExportToPDF(r Recipe, isBooklet bool) error {
	outputDir := "Printables"
	var pdf *gofpdf.Fpdf
	if isBooklet {
		pdf = gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "in", Size: gofpdf.SizeType{Wd: 5.5, Ht: 8.5}})
		pdf.SetMargins(0.75, 0.5, 0.5)
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
	pdf.AddPage() // Start the first page
	drawRecipePage(pdf, r, isBooklet, tr)

	suffix := ""
	if isBooklet {
		suffix = "_Booklet"
	}
	return pdf.OutputFileAndClose(filepath.Join(outputDir, r.Title+suffix+".pdf"))
}

func drawRecipePage(pdf *gofpdf.Fpdf, r Recipe, isBooklet bool, tr func(string) string) {
	// Standard Letter defaults
	startX, titleSize, rowH, sidebarW, contentW, gutter := 1.0, 32.0, 0.5, 2.2, 4.5, 0.3
	accentW, bodySize, headerSize := 0.4, 11.0, 13.0

	if isBooklet {
		startX, titleSize, rowH, sidebarW, contentW, gutter = 0.75, 20.0, 0.35, 1.4, 2.6, 0.2
		accentW, bodySize, headerSize = 0.2, 9.0, 11.0
	}

	// 1. Accent Bar
	pdf.SetFillColor(60, 60, 60)
	pdf.Rect(0, 0, accentW, 12, "F")

	// 2. Title Header
	pdf.SetX(startX)
	pdf.SetFont("Times", "B", titleSize)
	pdf.SetTextColor(40, 40, 40)
	pdf.MultiCell(0, rowH, tr(r.Title), "", "L", false)

	// 3. Tags
	pdf.SetX(startX)
	pdf.SetFont("Arial", "I", bodySize-1)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 0.2, tr(strings.Join(r.Tags, "  •  ")), "", 1, "L", false, 0, "")
	pdf.Ln(0.2)

	// Save the Y position after the header so columns align
	colY := pdf.GetY()

	// 4. Left Column: Ingredients
	pdf.SetX(startX)
	pdf.SetFont("Arial", "B", headerSize)
	pdf.SetTextColor(80, 20, 20)
	pdf.Cell(sidebarW, 0.35, tr("INGREDIENTS"))
	pdf.Ln(0.35)
	pdf.SetFont("Arial", "", bodySize)
	pdf.SetTextColor(0, 0, 0)
	for _, ing := range r.Ingredients {
		pdf.SetX(startX)
		pdf.MultiCell(sidebarW, 0.22, tr("• "+ing), "", "L", false)
		pdf.Ln(0.05)
	}

	// 5. Right Column: Preparation
	pdf.SetY(colY) // Reset height to match Ingredients header
	mainColX := startX + sidebarW + gutter
	pdf.SetX(mainColX)
	pdf.SetFont("Arial", "B", headerSize)
	pdf.SetTextColor(80, 20, 20)
	pdf.Cell(contentW, 0.35, tr("PREPARATION"))
	pdf.Ln(0.35)

	pdf.SetFont("Times", "", bodySize+1)
	pdf.SetTextColor(30, 30, 30)
	for i, step := range r.Instructions {
		pdf.SetX(mainColX)
		pdf.SetFont("Times", "B", bodySize+1)
		pdf.Cell(0.25, 0.25, fmt.Sprintf("%d. ", i+1))
		pdf.SetFont("Times", "", bodySize+1)
		pdf.MultiCell(contentW, 0.25, tr(step), "", "L", false)
		pdf.Ln(0.12)
	}

	// 6. Notes
	if r.Notes != "" {
		pdf.Ln(0.2)
		pdf.SetX(mainColX)
		pdf.SetFillColor(255, 255, 240)
		pdf.SetFont("Arial", "I", bodySize)
		pdf.MultiCell(contentW, 0.22, tr("Cook's Note: "+r.Notes), "L", "L", true)
	}
}
