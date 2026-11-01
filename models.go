// File: models.go
package main

// Recipe is the shared API shape for one dish: title, tags, ingredients,
// instructions, notes. The Elm decoder and the Flutter model both read this
// exact shape, so rename a field only together with those two clients.
type Recipe struct {
	Title        string   `json:"title"`
	Tags         []string `json:"tags"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	Notes        string   `json:"notes"`
}
