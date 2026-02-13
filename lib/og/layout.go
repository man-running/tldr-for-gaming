package og

import (
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const (
	imgWidth          = 2400
	imgHeight         = 1256
	paddingX          = 192 // 48 * 4 (from px-48)
	paddingY          = 144 // 36 * 4
	baseTitleMaxWidth = 1500
	titleMaxHeight    = 640
	lineHeightRatio   = 1.2
	avgCharWidthRatio = 0.60 // Conservative estimate for Lato
)

// computeTitleMetrics calculates the optimal font size and resulting block height for the title,
// given a maximum allowed text block width.
func computeTitleMetrics(text string, boldFont *truetype.Font, maxWidth float64) (fontSize, titleHeight float64) {
	length := float64(len(text))
	minLen, maxLen := 55.0, 140.0
	minSize, maxSize := 80.0, 200.0

	// 1. Length-aware interpolation
	clampedLength := math.Max(minLen, math.Min(length, maxLen))
	t := (clampedLength - minLen) / (maxLen - minLen)
	interpolatedSize := maxSize - t*(maxSize-minSize)

	// estimateHeight calculates the rendered height of the text block for a given font size.
	estimateHeight := func(size float64) float64 {
		charsPerLine := math.Max(1, math.Floor(maxWidth/(size*avgCharWidthRatio)))
		lines := math.Max(1, math.Ceil(float64(len(text))/charsPerLine))
		return lines * size * lineHeightRatio
	}

	// 2. Fit-to-box using binary search solver
	lo, hi := minSize, math.Min(maxSize, interpolatedSize)
	best := lo
	for i := 0; i < 16; i++ {
		mid := math.Floor((lo + hi) / 2)
		if estimateHeight(mid) <= titleMaxHeight {
			best = mid
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	finalFontSize := math.Max(minSize, math.Min(best, maxSize))
	finalTitleHeight := math.Min(estimateHeight(finalFontSize), titleMaxHeight)

	return finalFontSize, math.Floor(finalTitleHeight)
}

// RenderImage creates the OG image using the gg library.
func RenderImage(title string, bg image.Image, logo *oksvg.SvgIcon, boldFont, blackFont *truetype.Font) (*gg.Context, error) {
	dc := gg.NewContext(imgWidth, imgHeight)

	// Draw background and gradient
	dc.DrawImage(bg, 0, 0)
	grad := gg.NewLinearGradient(0, 0, 0, imgHeight)
	grad.AddColorStop(0, color.NRGBA{R: 0, G: 0, B: 0, A: 102})   // 0.40 alpha
	grad.AddColorStop(0.6, color.NRGBA{R: 0, G: 0, B: 0, A: 140}) // 0.55 alpha
	grad.AddColorStop(1, color.NRGBA{R: 0, G: 0, B: 0, A: 166})   // 0.65 alpha
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, imgWidth, imgHeight)
	dc.Fill()

	// Calculate title metrics with dynamic width that accounts for the logo area on the right.
	// Begin with a base width, then iteratively refine based on the logo size which depends on title height.
	var titleMaxWidth float64 = baseTitleMaxWidth
	var titleFontSize float64
	var titleHeight float64
	const rightGap = 64.0 // keep spacing between text block and logo

	// Initial computation with base width
	titleFontSize, titleHeight = computeTitleMetrics(title, boldFont, titleMaxWidth)

	for i := 0; i < 3; i++ {
		// Compute logo size from current title height
		craneHeight := math.Floor(titleHeight * 0.8)
		craneWidth := math.Floor((craneHeight * 4) / 3)

		// Available width for text is from left padding to the logo's left edge minus a gap
		logoX := float64(imgWidth-paddingX) - craneWidth
		titleX := float64(paddingX)
		availableWidth := math.Max(0, (logoX-rightGap)-titleX)

		// Constrain by the base max as an upper bound
		newMaxWidth := math.Min(float64(baseTitleMaxWidth), availableWidth)

		// If width hasn't changed meaningfully, stop
		if math.Abs(newMaxWidth-titleMaxWidth) < 1 {
			break
		}

		titleMaxWidth = newMaxWidth
		titleFontSize, titleHeight = computeTitleMetrics(title, boldFont, titleMaxWidth)
	}

	// Draw title
	dc.SetColor(color.White)
	titleFace := CreateFontFace(boldFont, titleFontSize)
	dc.SetFontFace(titleFace)
	// Position title block: vertically centered, horizontally at left padding
	titleX := float64(paddingX)
	titleY := (imgHeight - titleHeight) / 2
	dc.DrawStringWrapped(title, titleX, titleY, 0, 0, titleMaxWidth, lineHeightRatio, gg.AlignLeft)

	// Draw branding text below title
	blackFace := CreateFontFace(blackFont, 62)
	dc.SetFontFace(blackFace)
	brandY := titleY + titleHeight + 24 + 62                 // Add gap and font size for alignment
	dc.SetColor(color.NRGBA{R: 255, G: 255, B: 255, A: 217}) // white/85
	dc.DrawString("tldr.", titleX, brandY)

	// Measure "tldr." to position "takara.ai"
	tldrWidth, _ := dc.MeasureString("tldr.")
	takaraX := titleX + tldrWidth
	dc.SetColor(color.NRGBA{R: 0xD9, G: 0x10, B: 0x09, A: 255}) // #D91009
	dc.DrawString("takara.ai", takaraX, brandY)

	// Calculate logo size and draw it based on final title height
	craneHeight := math.Floor(titleHeight * 0.8)
	craneWidth := math.Floor((craneHeight * 4) / 3)
	logo.SetTarget(0, 0, craneWidth, craneHeight)
	logoImage := image.NewRGBA(image.Rect(0, 0, int(craneWidth), int(craneHeight)))
	scanner := rasterx.NewScannerGV(int(craneWidth), int(craneHeight), logoImage, logoImage.Bounds())
	rasterizer := rasterx.NewDasher(int(craneWidth), int(craneHeight), scanner)
	logo.Draw(rasterizer, 1.0)

	logoX := imgWidth - paddingX - int(craneWidth)
	logoY := (imgHeight - int(craneHeight)) / 2
	dc.DrawImage(logoImage, logoX, logoY)

	return dc, nil
}
