package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}
	return doc
}

func TestScrapeJSONLDHowToSteps(t *testing.T) {
	html := `<html><head><script type="application/ld+json">
{
  "@context": "https://schema.org/",
  "@type": "Recipe",
  "name": "Banana Bread",
  "description": "A moist classic.",
  "recipeCategory": "Dessert",
  "recipeCuisine": "American",
  "keywords": "quick, easy",
  "recipeYield": "1 loaf",
  "totalTime": "PT1H10M",
  "recipeIngredient": ["3 ripe bananas", "2 cups flour", "1 cup sugar"],
  "recipeInstructions": [
    {"@type": "HowToStep", "text": "Preheat oven to 350F."},
    {"@type": "HowToStep", "text": "Mix all ingredients."},
    {"@type": "HowToStep", "text": "Bake for 1 hour."}
  ]
}
</script></head><body></body></html>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/banana-bread")

	if r.Title != "Banana Bread" {
		t.Errorf("title = %q, want Banana Bread", r.Title)
	}
	if len(r.Ingredients) != 3 || r.Ingredients[0] != "3 ripe bananas" {
		t.Errorf("ingredients = %v", r.Ingredients)
	}
	if len(r.Instructions) != 3 || r.Instructions[2] != "Bake for 1 hour." {
		t.Errorf("instructions = %v", r.Instructions)
	}
	for _, want := range []string{"Dessert", "American", "quick", "easy", "Imported"} {
		if !contains(r.Tags, want) {
			t.Errorf("tags = %v, missing %q", r.Tags, want)
		}
	}
	for _, want := range []string{"Source: https://example.com/banana-bread", "A moist classic.", "Yield: 1 loaf", "Time: 1h 10m"} {
		if !strings.Contains(r.Notes, want) {
			t.Errorf("notes = %q, missing %q", r.Notes, want)
		}
	}
}

func TestScrapeJSONLDGraph(t *testing.T) {
	html := `<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@graph": [
    {"@type": "WebPage", "name": "Some page"},
    {"@type": "Recipe", "name": "Tomato Soup", "recipeIngredient": ["Tomatoes", "Water"], "recipeInstructions": ["Boil.", "Blend."]}
  ]
}
</script>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/soup")

	if r.Title != "Tomato Soup" {
		t.Errorf("title = %q, want Tomato Soup", r.Title)
	}
	if len(r.Ingredients) != 2 || len(r.Instructions) != 2 {
		t.Errorf("ingredients=%v instructions=%v", r.Ingredients, r.Instructions)
	}
}

func TestScrapeJSONLDPropertyValueIngredients(t *testing.T) {
	html := `<script type="application/ld+json">
{"@type":"Recipe","name":"Omelette",
 "recipeIngredient":["2 eggs",{"@type":"PropertyValue","value":1,"name":"pinch of salt"}],
 "recipeInstructions":"Beat eggs. Cook. Serve."}
</script>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/omelette")

	if len(r.Ingredients) != 2 || r.Ingredients[1] != "pinch of salt" {
		t.Errorf("ingredients = %v", r.Ingredients)
	}
	if len(r.Instructions) != 1 || r.Instructions[0] != "Beat eggs. Cook. Serve." {
		t.Errorf("instructions = %v", r.Instructions)
	}
}

func TestScrapeMicrodata(t *testing.T) {
	html := `<body>
<h1 itemprop="name">Pancakes</h1>
<ul>
  <li itemprop="recipeIngredient">1 cup flour</li>
  <li itemprop="recipeIngredient">1 egg</li>
</ul>
<ol itemprop="recipeInstructions">
  <li>Mix.</li>
  <li>Cook.</li>
</ol>
<span itemprop="recipeCategory">Breakfast</span>
<span itemprop="recipeYield">4 pancakes</span>
</body>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/pancakes")

	if r.Title != "Pancakes" {
		t.Errorf("title = %q, want Pancakes", r.Title)
	}
	if len(r.Ingredients) != 2 || len(r.Instructions) != 2 {
		t.Errorf("ingredients=%v instructions=%v", r.Ingredients, r.Instructions)
	}
	if !contains(r.Tags, "Breakfast") || !contains(r.Tags, "Imported") {
		t.Errorf("tags = %v", r.Tags)
	}
	if !strings.Contains(r.Notes, "Yield: 4 pancakes") {
		t.Errorf("notes = %q", r.Notes)
	}
}

func TestScrapeHeuristic(t *testing.T) {
	html := `<body>
<h1>Spaghetti</h1>
<ul class="ingredients-list">
  <li class="ingredient">Pasta</li>
  <li class="ingredient">Sauce</li>
</ul>
<ol class="instructions">
  <li class="step">Boil pasta.</li>
  <li class="step">Add sauce.</li>
</ol>
</body>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/spaghetti")

	if r.Title != "Spaghetti" {
		t.Errorf("title = %q", r.Title)
	}
	if len(r.Ingredients) != 2 || r.Ingredients[0] != "Pasta" {
		t.Errorf("ingredients = %v", r.Ingredients)
	}
	if len(r.Instructions) != 2 || r.Instructions[1] != "Add sauce." {
		t.Errorf("instructions = %v", r.Instructions)
	}
	if !contains(r.Tags, "Imported") {
		t.Errorf("tags = %v", r.Tags)
	}
}

func TestHumanizeDuration(t *testing.T) {
	cases := map[string]string{
		"PT1H10M": "1h 10m",
		"PT15M":   "15m",
		"PT2H":    "2h",
		"bogus":   "bogus",
	}
	for in, want := range cases {
		if got := humanizeDuration(in); got != want {
			t.Errorf("humanizeDuration(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeIngredients(t *testing.T) {
	in := []string{"- 1 cup sugar", "1 cup sugar", "• 2 eggs", "  flour  "}
	got := normalizeIngredients(in)
	want := []string{"1 cup sugar", "2 eggs", "flour"}
	if len(got) != len(want) {
		t.Fatalf("normalizeIngredients = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("normalizeIngredients[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScrapeJSONLDTimes(t *testing.T) {
	html := `<script type="application/ld+json">
{"@type":"Recipe","name":"Roast","recipeIngredient":["chicken"],
 "recipeInstructions":["Cook."],
 "prepTime":"PT15M","cookTime":"PT1H","totalTime":"PT1H15M"}
</script>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/roast")

	for _, want := range []string{"Prep: 15m", "Cook: 1h", "Time: 1h 15m"} {
		if !strings.Contains(r.Notes, want) {
			t.Errorf("notes = %q, missing %q", r.Notes, want)
		}
	}
}

func TestScrapeHowToSectionPrefix(t *testing.T) {
	html := `<script type="application/ld+json">
{"@type":"Recipe","name":"Pie","recipeIngredient":["dough"],
 "recipeInstructions":[
   {"@type":"HowToSection","name":"Crust","itemListElement":[
     {"@type":"HowToStep","text":"Roll dough."},
     {"@type":"HowToStep","text":"Bake crust."}
   ]},
   {"@type":"HowToStep","text":"Serve."}
 ]}
</script>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/pie")

	want := []string{"Crust: Roll dough.", "Crust: Bake crust.", "Serve."}
	if len(r.Instructions) != len(want) {
		t.Fatalf("instructions = %v, want %v", r.Instructions, want)
	}
	for i := range want {
		if r.Instructions[i] != want[i] {
			t.Errorf("instructions[%d] = %q, want %q", i, r.Instructions[i], want[i])
		}
	}
}

func TestHostStrategyOverride(t *testing.T) {
	old := hostStrategy["example.com"]
	hostStrategy["example.com"] = "heuristic"
	defer func() { hostStrategy["example.com"] = old }()

	html := `<body><h1>Heuristic Title</h1>
<script type="application/ld+json">
{"@type":"Recipe","name":"JSONLD Title","recipeIngredient":["x"],"recipeInstructions":["Y."]}
</script>
<ul><li class="ingredient">Pasta</li></ul>
<ol><li class="step">Boil.</li></ol></body>`

	r := scrapeRecipe(parseHTML(t, html), "https://example.com/recipe")

	if r.Title != "Heuristic Title" {
		t.Errorf("title = %q, want Heuristic Title (override should force heuristic)", r.Title)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestNewScrapeRequestSetsBrowserHeaders(t *testing.T) {
	req, err := newScrapeRequest("https://example.com/recipe")
	if err != nil {
		t.Fatalf("newScrapeRequest: %v", err)
	}
	if got := req.Header.Get("User-Agent"); got != scrapeUserAgent {
		t.Errorf("User-Agent = %q, want %q", got, scrapeUserAgent)
	}
	if req.Header.Get("Accept") == "" {
		t.Error("Accept header not set")
	}
}
