package main

type Recipe struct {
	Title        string
	Tags         []string
	Ingredients  []string
	Instructions []string
	Notes        string
	SourceFile   string // New field to track the filename
}
