package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ArtworkAPIResponse represents the structure of gopherize.me API response
type ArtworkAPIResponse struct {
	Categories []Category `json:"categories"`
}

type Category struct {
	ID     string  `json:"id"`     // e.g., "artwork/010-Body"
	Name   string  `json:"name"`   // e.g., "Body"
	Images []Image `json:"images"` // Array of images
}

type Image struct {
	ID           string `json:"id"`            // e.g., "artwork/010-Body/blue_gopher.png"
	Name         string `json:"name"`          // e.g., "blue gopher"
	Href         string `json:"href"`          // Download URL
	ThumbnailHref string `json:"thumbnail_href"` // Thumbnail URL
}

func main() {
	fmt.Println("Downloading gopherize.me artwork...")
	fmt.Println("Fetching artwork metadata from API...")

	// Fetch artwork metadata from gopherize.me API
	// According to the documentation, there's an API endpoint
	apiURL := "https://gopherize.me/api/artwork"
	
	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("Error: Could not fetch artwork metadata from API: %v\n", err)
		fmt.Println("\nAlternative: You can manually download artwork from:")
		fmt.Println("1. Visit https://gopherize.me")
		fmt.Println("2. Check the browser's network tab for artwork URLs")
		fmt.Println("3. Or contact the gopherize.me maintainers for artwork access")
		fmt.Println("\nArtwork should be organized in folders like:")
		fmt.Println("  assets/artwork/000-Body/Feature1.png")
		fmt.Println("  assets/artwork/010-Eyes/Feature1.png")
		fmt.Println("  assets/artwork/020-Mouth/Feature1.png")
		fmt.Println("  etc.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: API returned status %d\n", resp.StatusCode)
		fmt.Println("\nAlternative: You can manually download artwork from:")
		fmt.Println("1. Visit https://gopherize.me")
		fmt.Println("2. Check the browser's network tab for artwork URLs")
		fmt.Println("3. Or contact the gopherize.me maintainers for artwork access")
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading API response: %v\n", err)
		os.Exit(1)
	}

	var artwork ArtworkAPIResponse
	if err := json.Unmarshal(body, &artwork); err != nil {
		fmt.Printf("Error parsing API response: %v\n", err)
		fmt.Println("\nNote: The API format may have changed.")
		fmt.Println("You may need to manually download the artwork.")
		os.Exit(1)
	}

	// Create artwork directory
	artworkDir := "assets/artwork"
	if err := os.MkdirAll(artworkDir, 0755); err != nil {
		fmt.Printf("Error creating artwork directory: %v\n", err)
		os.Exit(1)
	}

	// Download each category and its images
	for _, category := range artwork.Categories {
		// Extract order number from category ID (e.g., "artwork/010-Body" -> 010)
		order, categoryName, err := parseCategoryID(category.ID)
		if err != nil {
			fmt.Printf("Warning: Could not parse category ID %s: %v\n", category.ID, err)
			// Fallback: use category name and try to extract order from ID
			order = extractOrderFromID(category.ID)
			categoryName = category.Name
		}

		// Create category folder with order prefix (e.g., "010-Body")
		categoryFolder := fmt.Sprintf("%03d-%s", order, categoryName)
		categoryPath := filepath.Join(artworkDir, categoryFolder)
		
		if err := os.MkdirAll(categoryPath, 0755); err != nil {
			fmt.Printf("Error creating category folder %s: %v\n", categoryFolder, err)
			continue
		}

		fmt.Printf("Downloading category: %s (%d images)...\n", categoryName, len(category.Images))

		// Download each image
		for i, image := range category.Images {
			// Extract filename from image ID or use name
			filename := extractFilename(image.ID, image.Name)
			filepath := filepath.Join(categoryPath, filename)

			// Skip if file already exists
			if _, err := os.Stat(filepath); err == nil {
				fmt.Printf("  [%d/%d] Skipping %s (already exists)\n", i+1, len(category.Images), filename)
				continue
			}

			// Download the image
			fmt.Printf("  [%d/%d] Downloading %s...\n", i+1, len(category.Images), filename)
			if err := downloadFile(image.Href, filepath); err != nil {
				fmt.Printf("  Error downloading %s: %v\n", filename, err)
				continue
			}
		}
	}

	fmt.Println("\nArtwork download complete!")
	fmt.Printf("Artwork saved to: %s\n", artworkDir)
}

// parseCategoryID extracts order number and category name from ID like "artwork/010-Body"
func parseCategoryID(id string) (int, string, error) {
	// Pattern: "artwork/010-Body" -> order=010, name="Body"
	re := regexp.MustCompile(`^artwork/(\d+)-(.+)$`)
	matches := re.FindStringSubmatch(id)
	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid category ID format: %s", id)
	}

	order, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", err
	}

	return order, matches[2], nil
}

// extractOrderFromID tries to extract order number from category ID as fallback
func extractOrderFromID(id string) int {
	re := regexp.MustCompile(`/(\d+)-`)
	matches := re.FindStringSubmatch(id)
	if len(matches) >= 2 {
		if order, err := strconv.Atoi(matches[1]); err == nil {
			return order
		}
	}
	return 0
}

// extractFilename extracts filename from image ID or generates from name
func extractFilename(imageID, imageName string) string {
	// Try to extract filename from ID (e.g., "artwork/010-Body/blue_gopher.png" -> "blue_gopher.png")
	if strings.Contains(imageID, "/") {
		parts := strings.Split(imageID, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			if strings.HasSuffix(filename, ".png") {
				return filename
			}
		}
	}

	// Fallback: generate filename from name (replace spaces with underscores, add .png)
	filename := strings.ReplaceAll(imageName, " ", "_") + ".png"
	return filename
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

