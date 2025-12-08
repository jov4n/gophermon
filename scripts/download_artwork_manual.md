# Downloading Gopherize.me Artwork

Since the gopherize.me artwork is stored in Google Cloud Storage and may not have a public API, you'll need to download it manually or use the gopherize.me website.

## Option 1: Using the gopherize.me Website

1. Visit https://gopherize.me in your browser
2. Open browser developer tools (F12)
3. Go to the Network tab
4. Generate a gopher on the website
5. Look for image requests - they should be from Google Cloud Storage
6. The URLs will follow a pattern like: `https://storage.googleapis.com/[bucket]/artwork/[category]/[feature].png`

## Option 2: Manual Structure

Create the following directory structure in `assets/artwork/`:

```
assets/artwork/
├── 000-Body/
│   ├── Feature1.png
│   ├── Feature2.png
│   └── ...
├── 010-Eyes/
│   ├── Feature1.png
│   ├── Feature2.png
│   └── ...
├── 020-Mouth/
│   ├── Feature1.png
│   └── ...
└── ... (more categories)
```

## Category Naming Rules

- Categories are numbered with 3 digits (000, 010, 020, etc.) for ordering
- Numbers are followed by a hyphen and the category name
- Example: `000-Body`, `010-Eyes`, `020-Mouth`, `030-Hat`, etc.

## Feature Naming Rules

- Features are PNG files
- Underscores in filenames become spaces in the UI (e.g., `Pirate_Beard.png` → "Pirate Beard")
- All images must be the same size
- Images are overlaid in category order to build the final gopher

## Testing

Once you have the artwork downloaded, test the generator:

```bash
go run cmd/bot/main.go
```

The bot will load all categories and use them to generate gophers!

