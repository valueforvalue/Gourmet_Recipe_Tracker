package main

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// hostStrategy forces a specific extraction strategy for a host. Empty by
// default (JSON-LD -> microdata -> heuristic). Add an entry when a site's
// markup breaks the default order, e.g. "broken.example": "heuristic".
var hostStrategy = map[string]string{}

// scrapeRecipe extracts a Recipe from an already-parsed HTML document. It
// tries schema.org/Recipe JSON-LD first, then microdata (itemprop), then
// CSS-class heuristics as a last resort, unless a host override says
// otherwise.
func scrapeRecipe(doc *goquery.Document, sourceURL string) Recipe {
	if strategy, ok := hostStrategy[hostOf(sourceURL)]; ok {
		switch strategy {
		case "jsonld":
			if r, ok := scrapeJSONLD(doc, sourceURL); ok {
				return r
			}
		case "microdata":
			if r, ok := scrapeMicrodata(doc, sourceURL); ok {
				return r
			}
		case "heuristic":
			return scrapeHeuristic(doc, sourceURL)
		}
	}
	if r, ok := scrapeJSONLD(doc, sourceURL); ok {
		return r
	}
	if r, ok := scrapeMicrodata(doc, sourceURL); ok {
		return r
	}
	return scrapeHeuristic(doc, sourceURL)
}

// hostOf returns the lowercase hostname of a URL, or "" when unparseable.
func hostOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

// --- JSON-LD (schema.org/Recipe) ---

// scrapeJSONLD finds the first schema.org/Recipe object across all JSON-LD
// script blocks (including objects nested in @graph) and extracts a Recipe.
func scrapeJSONLD(doc *goquery.Document, sourceURL string) (Recipe, bool) {
	var obj map[string]any
	doc.Find("script[type*='ld+json']").Each(func(_ int, s *goquery.Selection) {
		if obj != nil {
			return
		}
		raw := strings.TrimSpace(s.Text())
		if raw == "" {
			return
		}
		var data any
		if json.Unmarshal([]byte(raw), &data) != nil {
			return
		}
		obj = findRecipeObject(data)
	})
	if obj == nil {
		return Recipe{}, false
	}

	r := Recipe{
		Title:        cleanText(firstString(obj["name"])),
		Ingredients:  normalizeIngredients(ingredientsFromJSON(obj["recipeIngredient"])),
		Instructions: cleanStrings(instructionsFromJSON(obj["recipeInstructions"])),
		Tags:         tagsFromJSON(obj),
		Notes:        notesFromJSON(obj, sourceURL),
	}
	// A recipe block with neither ingredients nor instructions isn't useful;
	// fall through to the next strategy.
	if len(r.Ingredients) == 0 && len(r.Instructions) == 0 {
		return Recipe{}, false
	}
	return r, true
}

// findRecipeObject recursively searches a decoded JSON value for an object
// whose @type contains "Recipe".
func findRecipeObject(v any) map[string]any {
	switch t := v.(type) {
	case map[string]any:
		if typeIs(t["@type"], "recipe") {
			return t
		}
		for _, val := range t {
			if r := findRecipeObject(val); r != nil {
				return r
			}
		}
	case []any:
		for _, item := range t {
			if r := findRecipeObject(item); r != nil {
				return r
			}
		}
	}
	return nil
}

// typeIs reports whether v (a string or array of strings) contains target,
// case-insensitively.
func typeIs(v any, target string) bool {
	switch t := v.(type) {
	case string:
		return strings.Contains(strings.ToLower(t), strings.ToLower(target))
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok && strings.Contains(strings.ToLower(s), strings.ToLower(target)) {
				return true
			}
		}
	}
	return false
}

// firstString returns the first non-empty string value found in v, recursing
// into arrays and reading common object fields.
func firstString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		for _, item := range t {
			if s := firstString(item); s != "" {
				return s
			}
		}
	case map[string]any:
		for _, key := range []string{"name", "text", "@value", "value"} {
			if s, ok := t[key].(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// ingredientsFromJSON handles recipeIngredient as a string, []string, or
// []PropertyValue objects.
func ingredientsFromJSON(v any) []string {
	var out []string
	switch t := v.(type) {
	case string:
		out = append(out, t)
	case []any:
		for _, item := range t {
			switch it := item.(type) {
			case string:
				out = append(out, it)
			case map[string]any:
				if s := firstString(it); s != "" {
					out = append(out, s)
				}
			}
		}
	}
	return out
}

// instructionsFromJSON handles recipeInstructions as a string, []string,
// []HowToStep, or HowToSection (with nested itemListElement/steps).
func instructionsFromJSON(v any) []string {
	var out []string
	switch t := v.(type) {
	case string:
		for _, line := range strings.Split(t, "\n") {
			out = append(out, line)
		}
	case []any:
		for _, item := range t {
			out = append(out, stepFromJSON(item)...)
		}
	case map[string]any:
		out = append(out, stepFromJSON(t)...)
	}
	return out
}

func stepFromJSON(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case map[string]any:
		// HowToSection groups nested steps; prefix them with the section name.
		if typeIs(t["@type"], "HowToSection") {
			name := cleanText(firstString(t["name"]))
			var out []string
			for _, key := range []string{"itemListElement", "steps", "step"} {
				for _, step := range instructionsFromJSON(t[key]) {
					if name != "" {
						step = name + ": " + step
					}
					out = append(out, step)
				}
			}
			return out
		}
		// HowToStep carries "text" (fall back to "name").
		if s, ok := t["text"].(string); ok {
			return []string{s}
		}
		if s, ok := t["name"].(string); ok {
			return []string{s}
		}
	}
	return nil
}

// tagsFromJSON builds tags from recipeCategory, recipeCuisine, and keywords,
// always appending the "Imported" marker.
func tagsFromJSON(m map[string]any) []string {
	var tags []string
	tags = append(tags, splitComma(firstString(m["recipeCategory"]))...)
	tags = append(tags, splitComma(firstString(m["recipeCuisine"]))...)
	tags = append(tags, splitComma(firstString(m["keywords"]))...)
	tags = append(tags, "Imported")
	return uniqueClean(tags)
}

// notesFromJSON composes notes from description, yield, and total time, with
// the source URL always first.
func notesFromJSON(m map[string]any, sourceURL string) string {
	var parts []string
	parts = append(parts, "Source: "+sourceURL)
	if d := cleanText(firstString(m["description"])); d != "" {
		parts = append(parts, d)
	}
	if y := cleanText(firstString(m["recipeYield"])); y != "" {
		parts = append(parts, "Yield: "+y)
	}
	if prep := cleanText(firstString(m["prepTime"])); prep != "" {
		parts = append(parts, "Prep: "+humanizeDuration(prep))
	}
	if cook := cleanText(firstString(m["cookTime"])); cook != "" {
		parts = append(parts, "Cook: "+humanizeDuration(cook))
	}
	if t := cleanText(firstString(m["totalTime"])); t != "" {
		parts = append(parts, "Time: "+humanizeDuration(t))
	}
	return strings.Join(parts, "\n\n")
}

// --- Microdata (itemprop) ---

// scrapeMicrodata extracts a Recipe from schema.org microdata attributes.
func scrapeMicrodata(doc *goquery.Document, sourceURL string) (Recipe, bool) {
	r := Recipe{Title: cleanText(doc.Find("[itemprop='name']").First().Text())}

	doc.Find("[itemprop='recipeIngredient']").Each(func(_ int, s *goquery.Selection) {
		r.Ingredients = append(r.Ingredients, cleanText(s.Text()))
	})

	doc.Find("[itemprop='recipeInstructions']").Each(func(_ int, s *goquery.Selection) {
		if lis := s.Find("li"); lis.Length() > 0 {
			lis.Each(func(_ int, li *goquery.Selection) {
				r.Instructions = append(r.Instructions, cleanText(li.Text()))
			})
		} else {
			r.Instructions = append(r.Instructions, cleanText(s.Text()))
		}
	})

	var tags []string
	doc.Find("[itemprop='recipeCategory'], [itemprop='recipeCuisine'], [itemprop='keywords']").Each(func(_ int, s *goquery.Selection) {
		tags = append(tags, splitComma(cleanText(s.Text()))...)
	})
	tags = append(tags, "Imported")
	r.Tags = uniqueClean(tags)
	r.Ingredients = normalizeIngredients(r.Ingredients)

	var parts []string
	parts = append(parts, "Source: "+sourceURL)
	if y := cleanText(doc.Find("[itemprop='recipeYield']").First().Text()); y != "" {
		parts = append(parts, "Yield: "+y)
	}
	r.Notes = strings.Join(parts, "\n\n")

	if len(r.Ingredients) == 0 && len(r.Instructions) == 0 {
		return Recipe{}, false
	}
	return r, true
}

// --- Heuristic (CSS classes) ---

// scrapeHeuristic is the last-resort extractor: it walks <li> elements whose
// class (or a nearby ancestor's class) mentions ingredients/instructions.
func scrapeHeuristic(doc *goquery.Document, sourceURL string) Recipe {
	r := Recipe{
		Title: cleanText(doc.Find("h1").First().Text()),
		Tags:  []string{"Imported"},
		Notes: "Source: " + sourceURL,
	}
	doc.Find("li").Each(func(_ int, s *goquery.Selection) {
		if classHas(s, "ingredient") {
			r.Ingredients = append(r.Ingredients, cleanText(s.Text()))
		}
		if classHas(s, "instruction") || classHas(s, "direction") || classHas(s, "step") {
			r.Instructions = append(r.Instructions, cleanText(s.Text()))
		}
	})
	r.Ingredients = normalizeIngredients(r.Ingredients)
	r.Instructions = cleanStrings(r.Instructions)
	return r
}

// classHas reports whether the element or one of its two nearest ancestors
// has a class attribute containing needle.
func classHas(s *goquery.Selection, needle string) bool {
	for i := 0; i < 3; i++ {
		if c, ok := s.Attr("class"); ok && strings.Contains(strings.ToLower(c), needle) {
			return true
		}
		s = s.Parent()
		if s.Length() == 0 {
			return false
		}
	}
	return false
}

// --- Helpers ---

// cleanText collapses all whitespace to single spaces.
func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// cleanStrings trims/normalizes a list and drops empty entries.
func cleanStrings(list []string) []string {
	var out []string
	for _, s := range list {
		if s = cleanText(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// normalizeIngredients cautiously cleans ingredient strings: it strips a
// leading bullet and removes exact duplicates. It deliberately does not merge
// near-identical strings or parse amounts, which risks losing detail.
func normalizeIngredients(list []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range list {
		s := cleanText(raw)
		for _, b := range []string{"•", "·", "–", "—", "-", "*"} {
			if strings.HasPrefix(s, b+" ") {
				s = cleanText(strings.TrimPrefix(s, b+" "))
				break
			}
		}
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// splitComma splits a comma-delimited string into cleaned pieces.
func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = cleanText(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// uniqueClean dedupes and cleans a list, preserving order.
func uniqueClean(list []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range list {
		if s = cleanText(s); s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// humanizeDuration renders an ISO 8601 duration like "PT1H30M" as "1h 30m".
// Unparseable input is returned unchanged.
func humanizeDuration(s string) string {
	body := strings.TrimPrefix(strings.ToUpper(s), "PT")
	if body == s {
		body = s
	}
	var parts []string
	if i := strings.IndexByte(body, 'H'); i >= 0 {
		parts = append(parts, body[:i]+"h")
		body = body[i+1:]
	}
	if i := strings.IndexByte(body, 'M'); i >= 0 {
		parts = append(parts, body[:i]+"m")
	}
	out := strings.TrimSpace(strings.Join(parts, " "))
	if out == "" {
		return s
	}
	return out
}
