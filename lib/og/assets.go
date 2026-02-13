package og

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/srwiley/oksvg"
	"golang.org/x/image/font"
)

var (
	// In-memory caches for assets, initialized once per warm instance.
	assetsOnce    sync.Once
	cachedBg      image.Image
	cachedLogo    *oksvg.SvgIcon
	assetsLoadErr error

	fontsOnce       sync.Once
	cachedFontBold  *truetype.Font
	cachedFontBlack *truetype.Font
	fontsLoadErr    error
)

// LoadImageAndLogo loads the background image and SVG logo once.
func LoadImageAndLogo() (image.Image, *oksvg.SvgIcon, error) {
	assetsOnce.Do(func() {
		// Get base URL from environment variable
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			assetsLoadErr = fmt.Errorf("BASE_URL environment variable not set")
			return
		}

		// Load background image from web URL
		bgURL := fmt.Sprintf("%s/assets/og/background.jpg", baseURL)
		bgResp, err := http.Get(bgURL)
		if err != nil {
			assetsLoadErr = fmt.Errorf("failed to fetch background image from %s: %w", bgURL, err)
			return
		}
		defer func() {
			if cerr := bgResp.Body.Close(); cerr != nil && assetsLoadErr == nil {
				assetsLoadErr = fmt.Errorf("failed to close background image response body: %w", cerr)
			}
		}()

		if bgResp.StatusCode != http.StatusOK {
			assetsLoadErr = fmt.Errorf("bad status when fetching background image: %s", bgResp.Status)
			return
		}

		cachedBg, _, err = image.Decode(bgResp.Body)
		if err != nil {
			assetsLoadErr = fmt.Errorf("failed to decode background image: %w", err)
			return
		}

		// Load SVG logo from web URL
		logoURL := fmt.Sprintf("%s/assets/og/red_crane_vector.svg", baseURL)
		logoResp, err := http.Get(logoURL)
		if err != nil {
			assetsLoadErr = fmt.Errorf("failed to fetch logo from %s: %w", logoURL, err)
			return
		}
		defer func() {
			if cerr := logoResp.Body.Close(); cerr != nil && assetsLoadErr == nil {
				assetsLoadErr = fmt.Errorf("failed to close logo response body: %w", cerr)
			}
		}()

		if logoResp.StatusCode != http.StatusOK {
			assetsLoadErr = fmt.Errorf("bad status when fetching logo: %s", logoResp.Status)
			return
		}

		cachedLogo, err = oksvg.ReadIconStream(logoResp.Body)
		if err != nil {
			assetsLoadErr = fmt.Errorf("failed to parse logo svg: %w", err)
			return
		}
	})
	return cachedBg, cachedLogo, assetsLoadErr
}

// LoadFonts loads the bold and black fonts, preferring HTTP via BASE_URL then falling back to local files.
func LoadFonts() (*truetype.Font, *truetype.Font, error) {
	fontsOnce.Do(func() {
		var err error

		// 1) Attempt to load fonts over HTTP using BASE_URL
		if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
			boldURL := fmt.Sprintf("%s/fonts/lato/Lato-Bold.ttf", baseURL)
			blackURL := fmt.Sprintf("%s/fonts/lato/Lato-Black.ttf", baseURL)

			boldResp, boldErr := http.Get(boldURL)
			if boldErr == nil && boldResp != nil {
				defer func() { _ = boldResp.Body.Close() }()
			}
			blackResp, blackErr := http.Get(blackURL)
			if blackErr == nil && blackResp != nil {
				defer func() { _ = blackResp.Body.Close() }()
			}

			if boldErr == nil && blackErr == nil && boldResp.StatusCode == http.StatusOK && blackResp.StatusCode == http.StatusOK {
				if boldData, readBoldErr := io.ReadAll(boldResp.Body); readBoldErr == nil {
					if blackData, readBlackErr := io.ReadAll(blackResp.Body); readBlackErr == nil {
						cachedFontBold, err = truetype.Parse(boldData)
						if err == nil {
							cachedFontBlack, err = truetype.Parse(blackData)
							if err == nil {
								return
							}
						}
					}
				}
				// On any error, fall through to local fallback
			}
		}

		// 2) Fallback to local files. Try multiple likely paths.
		var boldData []byte
		for _, p := range []string{
			"public/fonts/lato/Lato-Bold.ttf",
			"./public/fonts/lato/Lato-Bold.ttf",
			"fonts/lato/Lato-Bold.ttf",
			"./fonts/lato/Lato-Bold.ttf",
			"/public/fonts/lato/Lato-Bold.ttf",
			"/fonts/lato/Lato-Bold.ttf",
		} {
			if data, readErr := os.ReadFile(p); readErr == nil {
				boldData = data
				break
			}
		}
		if len(boldData) == 0 {
			fontsLoadErr = fmt.Errorf("failed to read local bold font from known paths")
			return
		}

		var blackData []byte
		for _, p := range []string{
			"public/fonts/lato/Lato-Black.ttf",
			"./public/fonts/lato/Lato-Black.ttf",
			"fonts/lato/Lato-Black.ttf",
			"./fonts/lato/Lato-Black.ttf",
			"/public/fonts/lato/Lato-Black.ttf",
			"/fonts/lato/Lato-Black.ttf",
		} {
			if data, readErr := os.ReadFile(p); readErr == nil {
				blackData = data
				break
			}
		}
		if len(blackData) == 0 {
			fontsLoadErr = fmt.Errorf("failed to read local black font from known paths")
			return
		}

		cachedFontBold, err = truetype.Parse(boldData)
		if err != nil {
			fontsLoadErr = fmt.Errorf("failed to parse bold font: %w", err)
			return
		}

		cachedFontBlack, err = truetype.Parse(blackData)
		if err != nil {
			fontsLoadErr = fmt.Errorf("failed to parse black font: %w", err)
			return
		}
	})
	return cachedFontBold, cachedFontBlack, fontsLoadErr
}

// CreateFontFace is a helper to create a font.Face for drawing.
func CreateFontFace(ttFont *truetype.Font, size float64) font.Face {
	return truetype.NewFace(ttFont, &truetype.Options{
		Size:    size,
		Hinting: font.HintingFull,
	})
}
