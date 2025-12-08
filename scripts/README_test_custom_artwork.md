# Custom Artwork Test Script

This script allows you to manually specify which asset files to composite together, perfect for testing specific combinations like `glitch_gopher.png` as the body.

## Usage

Run from the project root:

```bash
go run scripts/test_custom_artwork.go [flags]
```

## Flags

- `-body`: Body asset file (e.g., `glitch_gopher.png`)
- `-eyes`: Eyes asset file
- `-shirt`: Shirt asset file
- `-hair`: Hair asset file
- `-facial`: Facial hair asset file
- `-glasses`: Glasses asset file
- `-hat`: Hat asset file
- `-extra`: Extra asset file(s) - can specify multiple with comma (e.g., `bowtie.png,watch.png`)
- `-output`: Output file name (default: `test_custom.png`)
- `-assets`: Assets directory path (default: `assets/artwork`)
- `-list`: List all available assets by category

## Examples

### List Available Assets
```bash
go run scripts/test_custom_artwork.go -list
```

### Test Glitch Gopher Body
```bash
go run scripts/test_custom_artwork.go -body=glitch_gopher.png -eyes=crazy_eyes.png -output=test_glitch.png
```

### Full Custom Combination
```bash
go run scripts/test_custom_artwork.go \
  -body=glitch_gopher.png \
  -eyes=crazy_eyes.png \
  -hair=rainbow_hair.png \
  -glasses=sunglasses.png \
  -extra=bowtie.png,watch.png \
  -output=my_custom.png
```

### Just Body and Eyes
```bash
go run scripts/test_custom_artwork.go -body=blue_gopher.png -eyes=goofy_eyes.png
```

### Multiple Extras
```bash
go run scripts/test_custom_artwork.go \
  -body=green_gopher.png \
  -eyes=eyes.png \
  -extra=RedStapler.png,bowtie.png,camera.png \
  -output=extras_test.png
```

## File Naming

You can specify files in several ways:

1. **Just filename** (recommended): `glitch_gopher.png`
   - Script will search in appropriate category folders
   
2. **With category folder**: `010-Body/glitch_gopher.png`
   - More explicit, but usually not needed

3. **Full path**: `/full/path/to/glitch_gopher.png`
   - Use if file is outside the assets directory

## Layer Order

Layers are composited in this order (later layers appear on top):
1. Body (base layer)
2. Eyes
3. Shirt
4. Hair
5. Facial Hair
6. Glasses
7. Hat
8. Extras (all extras are composited last, in the order specified)

## Tips

- Use `-list` to see all available assets in each category
- You don't need to specify all layers - just the ones you want to test
- Extras are composited on top of everything else
- The script will warn you if it can't find a specified asset file

## Automatic Resizing

The script automatically resizes all layers to match the standard body size (1300x1392 pixels):
- **Body layer**: Resized to 1300x1392 if it's a different size (e.g., `glitch_gopher.png` is 1024x1024 and will be resized)
- **Other layers**: Automatically resized to match the body dimensions

This ensures all assets composite correctly regardless of their original dimensions.

## Output

The script generates a PNG file with:
- All specified layers composited together
- Proper layer ordering (body first, extras last)
- Standard dimensions: 1300x1392 pixels (matching standard body assets)

