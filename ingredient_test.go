package main

import (
	"math/big"
	"testing"
)

func TestParseIngredient(t *testing.T) {
	cases := []struct {
		in       string
		qty      string
		unit     string
		name     string
		scalable bool
	}{
		{"2 cups flour", "2", "cups", "flour", true},
		{"1/2 cup sugar", "1/2", "cup", "sugar", true},
		{"1 1/2 cups flour", "1 1/2", "cups", "flour", true},
		{"2 eggs", "2", "", "eggs", true},
		{"3 cloves garlic", "3", "cloves", "garlic", true},
		{"0.5 cup milk", "1/2", "cup", "milk", true},
		{"1½ cups flour", "1 1/2", "cups", "flour", true},
		{"¾ cup sugar", "3/4", "cup", "sugar", true},
		{"salt to taste", "", "", "salt to taste", false},
		{"a pinch of salt", "", "", "a pinch of salt", false},
		{"2-3 cloves garlic", "", "", "2-3 cloves garlic", false},
		{"1 (15 oz) can tomatoes", "", "", "1 (15 oz) can tomatoes", false},
	}

	for _, c := range cases {
		p := parseIngredient(c.in)
		if p.ok != c.scalable {
			t.Errorf("parseIngredient(%q).ok = %v, want %v", c.in, p.ok, c.scalable)
			continue
		}
		if !c.scalable {
			continue
		}
		if got := formatRat(p.qty); got != c.qty {
			t.Errorf("parseIngredient(%q) qty = %q, want %q", c.in, got, c.qty)
		}
		if p.unit != c.unit {
			t.Errorf("parseIngredient(%q) unit = %q, want %q", c.in, p.unit, c.unit)
		}
		if p.name != c.name {
			t.Errorf("parseIngredient(%q) name = %q, want %q", c.in, p.name, c.name)
		}
	}
}

func TestScaleIngredient(t *testing.T) {
	cases := []struct {
		in     string
		factor *big.Rat
		want   string
		ok     bool
	}{
		{"2 cups flour", big.NewRat(2, 1), "4 cups flour", true},
		{"1/2 cup sugar", big.NewRat(3, 1), "1 1/2 cup sugar", true},
		{"2 eggs", big.NewRat(3, 1), "6 eggs", true},
		{"1 1/2 cups flour", big.NewRat(2, 1), "3 cups flour", true},
		{"¾ cup sugar", big.NewRat(2, 1), "1 1/2 cup sugar", true},
		{"salt to taste", big.NewRat(2, 1), "salt to taste", false},
	}

	for _, c := range cases {
		got, ok := scaleIngredient(c.in, c.factor)
		if ok != c.ok || got != c.want {
			t.Errorf("scaleIngredient(%q, %v) = %q, %v; want %q, %v", c.in, c.factor, got, ok, c.want, c.ok)
		}
	}
}

func TestFormatRat(t *testing.T) {
	cases := map[*big.Rat]string{
		big.NewRat(1, 2): "1/2",
		big.NewRat(3, 2): "1 1/2",
		big.NewRat(2, 1): "2",
		big.NewRat(1, 3): "1/3",
		big.NewRat(5, 2): "2 1/2",
	}
	for in, want := range cases {
		if got := formatRat(in); got != want {
			t.Errorf("formatRat(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestScanRecipes(t *testing.T) {
	GlobalConfig.GitSync = false
	db := newSoftDeleteTestDB(t)

	for _, r := range []Recipe{
		{Title: "Scalable", Ingredients: []string{"2 cups flour"}},
		{Title: "Needs Fix", Ingredients: []string{"salt to taste"}},
	} {
		if err := saveRecipeDB(db, r); err != nil {
			t.Fatal(err)
		}
	}

	issues, err := scanUnscalable(db)
	if err != nil {
		t.Fatalf("scanUnscalable: %v", err)
	}
	if len(issues) != 1 || issues[0].Title != "Needs Fix" || issues[0].Ingredient != "salt to taste" {
		t.Errorf("issues = %+v", issues)
	}
}
