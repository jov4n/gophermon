# Artwork Generation Test Script

This script generates gopher artwork/sprites only, allowing you to test different asset combinations and manually mix/match layers.

## Usage

Run from the project root:

```bash
go run scripts/test_artwork_generation.go [flags]
```

## Flags

- `-rarity`: Gopher rarity
  - Options: `COMMON`, `UNCOMMON`, `RARE`, `EPIC`, `LEGENDARY`
  - If not specified: Random (weighted distribution)
  - Example: `-rarity=LEGENDARY`

- `-complexity`: Complexity score
  - Range: `1-15`
  - If not specified: Based on rarity (or random if rarity not specified)
  - Example: `-complexity=8`

- `-seed`: Random seed for reproducible generation
  - Example: `-seed=12345`
  - If not specified: Uses current time

- `-output`: Output file name (default: `test_artwork.png`)
  - Example: `-output=my_sprite.png`
  - If generating multiple, files will be numbered: `output_1.png`, `output_2.png`, etc.

- `-count`: Number of sprites to generate (default: `1`)
  - Example: `-count=5`

- `-info`: Show generation info (default: `true`)
  - Set to `false` to only show filenames
  - Example: `-info=false`

## Examples

### Random Artwork
```bash
go run scripts/test_artwork_generation.go
```

### Specific Rarity
```bash
go run scripts/test_artwork_generation.go -rarity=LEGENDARY
```

### Specific Complexity
```bash
go run scripts/test_artwork_generation.go -complexity=8
```

### Custom Output File
```bash
go run scripts/test_artwork_generation.go -output=my_sprite.png
```

### Generate Multiple Sprites
```bash
# Generate 5 random sprites
go run scripts/test_artwork_generation.go -count=5

# Generate 10 rare sprites
go run scripts/test_artwork_generation.go -count=10 -rarity=RARE
```

### Reproducible Generation
```bash
# Same seed = same sprite
go run scripts/test_artwork_generation.go -seed=12345
```

### Combine Options
```bash
go run scripts/test_artwork_generation.go \
  -rarity=EPIC \
  -complexity=7 \
  -seed=999 \
  -output=epic_sprite.png \
  -count=3
```

## Output

The script generates PNG sprite files with:

- **Full sprite image**: Complete gopher artwork
- **Layer information**: Shows which layers were used (in info mode)
- **Dimensions**: Image size
- **Rarity and complexity**: Generation parameters

## Use Cases

- **Test asset combinations**: Generate sprites with different rarities/complexities to see layer combinations
- **Batch generation**: Create multiple sprites at once for testing
- **Reproducible testing**: Use seeds to regenerate the same sprite
- **Manual mixing**: Generate sprites, then manually combine layers in your image editor

## Notes

- Complexity affects which layers are selected (higher = more layers)
- Rarity affects the pool of available layers
- Each sprite shows the layer list in the output (when `-info=true`)
- Multiple sprites are automatically numbered: `output_1.png`, `output_2.png`, etc.

