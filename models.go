// File: models.go
package main

type Recipe struct {
	Title        string   `json:"title"`
	Tags         []string `json:"tags"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	Notes        string   `json:"notes"`
}
