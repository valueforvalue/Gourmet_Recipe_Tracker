# Gourmet Recipe Tracker

A lightweight Go-based utility that transforms plain-text Notepad files into a searchable SQLite database and professionally formatted "Gourmet Edition" PDFs.

## How It Works
1. **Input**: Write recipes in a simple text format using the provided `Template.txt`.
2. **Process**: The Go engine parses the text, cleans up formatting (like double-numbering from copy-pastes), and handles character encoding for symbols like degrees (°).
3. **Storage**: Each recipe is indexed in a SQLite database (`recipes.db`) using the filename as a unique identifier to prevent duplicates.
4. **Output**: A two-column, "Cookbook-style" PDF is generated in the `/Printables` folder.

## Getting Started

### Prerequisites
- [Go](https://go.dev/dl/) (version 1.21 or higher)
- VS Code (recommended)

### Installation & Build
1. Clone the repository:
   ```bash
   git clone [your-private-repo-link]
   cd recipe-tracker

   go get modernc.org/sqlite
go get [github.com/jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf)

go build -o RecipeTracker.exe .

RECIPE: Bread Name
TAGS: Baking, Homemade

--- INGREDIENTS ---
- 5 cups flour
- 2 tsp salt

--- INSTRUCTIONS ---
1. Mix dry ingredients.
2. Let rise for 1 hour.

NOTES: Best served warm with butter.