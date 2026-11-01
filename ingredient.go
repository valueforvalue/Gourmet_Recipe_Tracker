package main

import (
	"database/sql"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// Ingredient parsing + scaling.
//
// Free-text ingredient lines ("2 cups flour", "1 1/2 tsp salt", "salt to
// taste") are parsed into a quantity (rational), an optional unit, and a
// name. Lines with no leading quantity, or with parenthetical notes, are
// deliberately treated as NOT scalable so the scanner can flag them for
// manual review.

type parsedIngredient struct {
	qty  *big.Rat
	unit string
	name string
	ok   bool // false => no safe quantity; do not scale
}

var (
	unicodeFrac = map[string]*big.Rat{
		"½": big.NewRat(1, 2),
		"⅓": big.NewRat(1, 3),
		"⅔": big.NewRat(2, 3),
		"¼": big.NewRat(1, 4),
		"¾": big.NewRat(3, 4),
		"⅕": big.NewRat(1, 5),
		"⅖": big.NewRat(2, 5),
		"⅗": big.NewRat(3, 5),
		"⅘": big.NewRat(4, 5),
		"⅛": big.NewRat(1, 8),
		"⅜": big.NewRat(3, 8),
		"⅝": big.NewRat(5, 8),
		"⅞": big.NewRat(7, 8),
	}
	fracRe   = regexp.MustCompile(`^(\d+)/(\d+)$`)
	decRe    = regexp.MustCompile(`^(\d+)\.(\d+)$`)
	intRe    = regexp.MustCompile(`^(\d+)$`)
	gluedRe  = regexp.MustCompile(`^(\d+)?([½⅓⅔¼¾⅕⅖⅗⅘⅛⅜⅝⅞])$`)
)

var unitSet = map[string]bool{
	"cup": true, "cups": true,
	"tablespoon": true, "tablespoons": true, "tbsp": true,
	"teaspoon": true, "teaspoons": true, "tsp": true,
	"g": true, "gram": true, "grams": true,
	"kg": true, "kilogram": true, "kilograms": true,
	"oz": true, "ounce": true, "ounces": true,
	"lb": true, "lbs": true, "pound": true, "pounds": true,
	"ml": true, "milliliter": true, "milliliters": true,
	"l": true, "liter": true, "liters": true,
	"clove": true, "cloves": true,
	"can": true, "cans": true,
	"package": true, "packages": true, "pkg": true,
	"slice": true, "slices": true,
	"piece": true, "pieces": true,
	"pinch": true, "pinches": true,
	"dash": true, "dashes": true,
	"bunch": true, "bunches": true,
	"stalk": true, "stalks": true,
	"sprig": true, "sprigs": true,
	"drop": true, "drops": true,
	"pint": true, "pints": true,
	"quart": true, "quarts": true,
	"gallon": true, "gallons": true,
}

// parseIngredient parses one free-text ingredient line.
func parseIngredient(line string) parsedIngredient {
	line = cleanText(line)
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return parsedIngredient{name: line}
	}

	qty, consumed, ok := parseQuantityAt(fields)
	if !ok {
		return parsedIngredient{name: line}
	}
	rest := fields[consumed:]

	var unit string
	if len(rest) > 0 && isUnit(rest[0]) {
		unit = strings.ToLower(strings.TrimSuffix(rest[0], "."))
		rest = rest[1:]
	}

	name := cleanText(strings.Join(rest, " "))
	// A parenthetical note ("1 (15 oz) can") is too ambiguous to auto-scale.
	if strings.ContainsAny(name, "()") {
		return parsedIngredient{name: line}
	}
	return parsedIngredient{qty: qty, unit: unit, name: name, ok: true}
}

// parseQuantityAt consumes a leading quantity (integer, decimal, fraction,
// mixed number, or glued Unicode fraction) and returns it plus how many
// fields were consumed.
func parseQuantityAt(fields []string) (*big.Rat, int, bool) {
	if len(fields) == 0 {
		return nil, 0, false
	}
	// Glued Unicode fraction: "1½".
	if m := gluedRe.FindStringSubmatch(fields[0]); m != nil {
		r := new(big.Rat)
		if m[1] != "" {
			n, _ := new(big.Int).SetString(m[1], 10)
			r.SetInt(n)
		}
		if f, ok := unicodeFrac[m[2]]; ok {
			r.Add(r, f)
		}
		return r, 1, true
	}

	q, ok := parseQuantityToken(fields[0])
	if !ok {
		return nil, 0, false
	}
	consumed := 1

	// Mixed number: "1 1/2" (integer followed by a fraction token).
	if len(fields) > 1 && strings.Contains(fields[1], "/") {
		if f, ok := parseQuantityToken(fields[1]); ok {
			q.Add(q, f)
			consumed = 2
		}
	}
	return q, consumed, true
}

func parseQuantityToken(tok string) (*big.Rat, bool) {
	if r, ok := unicodeFrac[tok]; ok {
		return new(big.Rat).Set(r), true
	}
	if m := fracRe.FindStringSubmatch(tok); m != nil {
		n, _ := new(big.Int).SetString(m[1], 10)
		d, _ := new(big.Int).SetString(m[2], 10)
		if d.Sign() == 0 {
			return nil, false
		}
		return new(big.Rat).SetFrac(n, d), true
	}
	if intRe.MatchString(tok) || decRe.MatchString(tok) {
		r := new(big.Rat)
		if _, ok := r.SetString(tok); ok {
			return r, true
		}
	}
	return nil, false
}

func isUnit(tok string) bool {
	return unitSet[strings.ToLower(strings.TrimSuffix(tok, "."))]
}

// isScalable reports whether a line can be safely scaled.
func isScalable(line string) bool {
	return parseIngredient(line).ok
}

// scaleIngredient multiplies the quantity in line by factor, returning the
// scaled line and whether the line was scalable. Unscalable lines are
// returned unchanged with ok=false.
func scaleIngredient(line string, factor *big.Rat) (string, bool) {
	p := parseIngredient(line)
	if !p.ok {
		return line, false
	}
	q := new(big.Rat).Mul(p.qty, factor)
	var parts []string
	parts = append(parts, formatRat(q))
	if p.unit != "" {
		parts = append(parts, p.unit)
	}
	if p.name != "" {
		parts = append(parts, p.name)
	}
	return strings.Join(parts, " "), true
}

// scaleRecipe returns a copy of r with each scalable ingredient multiplied by
// factor. Unscalable ingredients are left unchanged.
func scaleRecipe(r Recipe, factor *big.Rat) Recipe {
	out := r
	out.Ingredients = make([]string, len(r.Ingredients))
	for i, ing := range r.Ingredients {
		if s, ok := scaleIngredient(ing, factor); ok {
			out.Ingredients[i] = s
		} else {
			out.Ingredients[i] = ing
		}
	}
	return out
}

// ScanIssue flags one ingredient that can't be scaled.
type ScanIssue struct {
	Title      string
	Ingredient string
}

// scanUnscalable returns every ingredient that can't be safely scaled.
func scanUnscalable(db *sql.DB) ([]ScanIssue, error) {
	recipes, err := GetAllRecipes(db)
	if err != nil {
		return nil, err
	}
	var issues []ScanIssue
	for _, r := range recipes {
		for _, ing := range r.Ingredients {
			if !isScalable(ing) {
				issues = append(issues, ScanIssue{Title: r.Title, Ingredient: ing})
			}
		}
	}
	return issues, nil
}

// formatRat renders a rational as a clean quantity: integer, fraction, or
// mixed number ("2", "1/2", "1 1/2"). Falls back to a 2-decimal float for
// awkward denominators.
func formatRat(r *big.Rat) string {
	if !r.IsInt() {
		g := new(big.Int).GCD(nil, nil, r.Num(), r.Denom())
		num := new(big.Int).Quo(r.Num(), g)
		den := new(big.Int).Quo(r.Denom(), g)
		if den.Cmp(big.NewInt(100)) < 0 {
			if num.Cmp(den) >= 0 {
				whole := new(big.Int).Quo(num, den)
				rem := new(big.Int).Rem(num, den)
				if rem.Sign() == 0 {
					return whole.String()
				}
				return whole.String() + " " + rem.String() + "/" + den.String()
			}
			return num.String() + "/" + den.String()
		}
	}
	f, _ := r.Float64()
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(f, 'f', 2, 64), "0"), ".")
}
