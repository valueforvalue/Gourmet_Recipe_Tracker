# Gourmet Recipe Tracker v2.4

An interactive Go-based utility for managing a digital recipe library and generating professional North American Letter and Booklet-style PDFs.

## Core Features
- **Interactive ASCII Menu**: Manage your library, sync files, and generate cookbooks from a single console.
- **Dual-Mode Sync**: Generate individual recipes in **Standard Letter (8.5" x 11")** or **Booklet (5.5" x 8.5")**.
- **Master Cookbook Generation**: Compiles selected recipes into a single, indexed PDF with a clickable Table of Contents.
- **Config-Based Selection**: Automatically generates `cookbook_config.txt` to let you choose exactly which recipes to include in a master export.
- **SQLite Backend**: Maintains a persistent database of all recipes to prevent duplicates and allow for quick searching.

## Usage Instructions

### 1. Adding Recipes
Place your recipe `.txt` files in the `recipes_to_import/` folder. Use the provided `Template.txt` for consistent formatting.

### 2. Running the Program
Run `RecipeTracker.exe` and select from the following options:
- **[1] & [2]**: Syncs the folder to the database and generates individual PDFs.
- **[4] Master Cookbook**: Generates a list of all recipes. You can delete the ones you don't want, then the program builds a single Master PDF.

### 3. Printing for Physical Use (The "2-Up" Method)
To print Booklet-sized pages (5.5" x 8.5") efficiently onto standard paper:
- Open the PDF and select **Print**.
- Set Orientation to **Landscape**.
- Set **Multiple Pages per Sheet** to **2**.
- The **0.75" Left Margin** ensures space for hole-punching after you cut the sheet in half at the 5.5" mark.

## Build Information

### Prerequisites
- [Go](https://go.dev/dl/) (v1.21+)
- Dependencies:
  ```bash
  go get modernc.org/sqlite
  go get [github.com/jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf)