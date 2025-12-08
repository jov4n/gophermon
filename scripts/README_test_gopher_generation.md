# Gopher Generation Test Script

This script allows you to generate gophers with specific or random attributes for testing purposes.

## Usage

Run from the project root:

```bash
go run scripts/test_gopher_generation.go [flags]
```

## Flags

All flags are optional. If not specified, values will be randomly generated.

- `-archetype`: Gopher archetype
  - Options: `Hacker`, `Tank`, `Speedy`, `Support`, `Mage`
  - Example: `-archetype=Hacker`

- `-rarity`: Gopher rarity
  - Options: `COMMON`, `UNCOMMON`, `RARE`, `EPIC`, `LEGENDARY`
  - Example: `-rarity=LEGENDARY`

- `-evolution`: Evolution stage
  - Options: `0` (base), `1` (first evolution), `2` (final evolution)
  - Example: `-evolution=2`

- `-level`: Gopher level
  - Range: `1-100`
  - Example: `-level=50`

- `-complexity`: Complexity score
  - Range: `1-15`
  - If not specified, will be determined by rarity
  - Example: `-complexity=8`

- `-seed`: Random seed for reproducible generation
  - Example: `-seed=12345`

- `-output`: Save sprite to file (PNG)
  - Example: `-output=test_gopher.png`

- `-abilities`: Show abilities list (default: `true`)
  - Example: `-abilities=false` to hide

- `-stats`: Show stats (default: `true`)
  - Example: `-stats=false` to hide

- `-status`: Show status effects info (default: `true`)
  - Example: `-status=false` to hide

## Examples

### Random Gopher
```bash
go run scripts/test_gopher_generation.go
```

### Specific Archetype and Rarity
```bash
go run scripts/test_gopher_generation.go -archetype=Hacker -rarity=LEGENDARY
```

### Evolution Stage 2 Gopher
```bash
go run scripts/test_gopher_generation.go -evolution=2 -level=40
```

### Save Sprite to File
```bash
go run scripts/test_gopher_generation.go -output=test_gopher.png
```

### Full Custom Gopher
```bash
go run scripts/test_gopher_generation.go \
  -archetype=Mage \
  -rarity=EPIC \
  -evolution=1 \
  -level=25 \
  -complexity=7 \
  -seed=12345 \
  -output=my_gopher.png
```

### Test Legendary Abilities
```bash
go run scripts/test_gopher_generation.go -rarity=LEGENDARY -level=50
```

### Test Evolution Abilities
```bash
# Stage 1 abilities
go run scripts/test_gopher_generation.go -evolution=1 -level=20

# Stage 2 abilities
go run scripts/test_gopher_generation.go -evolution=2 -level=35
```

## Output

The script displays:

1. **Gopher Information**
   - Name, Archetype, Rarity
   - Evolution Stage, Level, Complexity
   - Primary and Secondary Types

2. **Stats** (if `-stats=true`)
   - HP, Attack, Defense, Speed

3. **Abilities** (if `-abilities=true`)
   - List of all available abilities
   - Shows which are evolution-specific or legendary
   - Power, Cost, and Targeting for each ability

4. **Status Effects Info** (if `-status=true`)
   - Explanation of all status effects

5. **Sprite** (if `-output` specified)
   - Saves the generated sprite as PNG

## Testing Specific Combinations

Use this script to test:

- **Evolution-specific abilities**: Set `-evolution=1` or `-evolution=2`
- **Legendary abilities**: Set `-rarity=LEGENDARY`
- **Type combinations**: Different archetypes have different ability sets
- **Status effects**: Generate gophers and check which abilities apply status effects
- **Ability progression**: Test how abilities unlock at different levels and evolution stages

## Notes

- Evolution stage is automatically limited by level (stage 1 requires level 16+, stage 2 requires level 32+)
- Legendary gophers get special legendary abilities
- Higher evolution stages unlock more powerful abilities
- The number of abilities shown depends on level and evolution stage

