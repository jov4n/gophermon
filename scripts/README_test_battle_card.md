# Battle Card Alignment Test Script

This script helps you manually align gophers and text on the battle screen.

## Usage

Run the script from the project root:

```bash
go run scripts/test_battle_card.go
```

This will generate `test_battle_card.png` in the project root.

## Adjusting Positions

Edit the `config` struct in `scripts/test_battle_card.go` to adjust:

- **MaxGopherSize**: Size of gophers (default: 350)
- **PlayerPlatformX/Y**: Position of player gopher platform (left side)
- **EnemyPlatformX/Y**: Position of enemy gopher platform (right side)
- **PlayerTextX/Y**: Position of player name/level text (left frame)
- **EnemyTextX/Y**: Position of enemy name/level text (right frame)

The script will print the calculated positions to help you fine-tune them.

## Example Output

The script outputs:
- Battle screen dimensions
- Calculated platform positions
- Calculated text positions
- Location of saved test image

After adjusting values, run the script again to see the updated alignment.

