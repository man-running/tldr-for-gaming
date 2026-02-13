package rendering

import (
	"image"
	"image/color"
	"io"
	"math"

	"github.com/HugoSmits86/nativewebp"
)

var colormap [256]color.NRGBA

func init() {
	// Baked-in coolwarm colormap from matplotlib
	colors := [][]int{
		{58, 76, 192}, {59, 77, 193}, {60, 79, 195}, {62, 81, 196}, {63, 83, 198}, {64, 84, 199}, {65, 86, 201}, {66, 88, 202}, {67, 90, 204}, {69, 91, 205}, {70, 93, 207}, {71, 95, 208}, {72, 96, 209}, {73, 98, 211}, {75, 100, 212}, {76, 102, 214}, {77, 103, 215}, {78, 105, 216}, {80, 107, 218}, {81, 108, 219}, {82, 110, 220}, {83, 112, 221}, {85, 113, 222}, {86, 115, 224}, {87, 117, 225}, {88, 118, 226}, {90, 120, 227}, {91, 121, 228}, {92, 123, 229}, {93, 125, 230}, {95, 126, 231}, {96, 128, 232}, {97, 130, 234}, {99, 131, 234}, {100, 133, 235}, {101, 134, 236}, {103, 136, 237}, {104, 137, 238}, {105, 139, 239}, {107, 141, 240}, {108, 142, 241}, {109, 144, 241}, {111, 145, 242}, {112, 147, 243}, {113, 148, 244}, {115, 149, 244}, {116, 151, 245}, {117, 152, 246}, {119, 154, 246}, {120, 155, 247}, {122, 157, 248}, {123, 158, 248}, {124, 160, 249}, {126, 161, 249}, {127, 162, 250}, {128, 164, 250}, {130, 165, 251}, {131, 166, 251}, {133, 168, 251}, {134, 169, 252}, {135, 170, 252}, {137, 172, 252}, {138, 173, 253}, {139, 174, 253}, {141, 175, 253}, {142, 177, 253}, {144, 178, 254}, {145, 179, 254}, {146, 180, 254}, {148, 181, 254}, {149, 183, 254}, {151, 184, 254}, {152, 185, 254}, {153, 186, 254}, {155, 187, 254}, {156, 188, 254}, {157, 189, 254}, {159, 190, 254}, {160, 191, 254}, {162, 192, 254}, {163, 193, 254}, {164, 194, 254}, {166, 195, 253}, {167, 196, 253}, {168, 197, 253}, {170, 198, 253}, {171, 199, 252}, {172, 200, 252}, {174, 201, 252}, {175, 202, 251}, {176, 203, 251}, {178, 203, 251}, {179, 204, 250}, {180, 205, 250}, {182, 206, 249}, {183, 207, 249}, {184, 207, 248}, {185, 208, 248}, {187, 209, 247}, {188, 209, 246}, {189, 210, 246}, {190, 211, 245}, {192, 211, 245}, {193, 212, 244}, {194, 212, 243}, {195, 213, 242}, {197, 213, 242}, {198, 214, 241}, {199, 214, 240}, {200, 215, 239}, {201, 215, 238}, {202, 216, 238}, {204, 216, 237}, {205, 217, 236}, {206, 217, 235}, {207, 217, 234}, {208, 218, 233}, {209, 218, 232}, {210, 218, 231}, {211, 219, 230}, {213, 219, 229}, {214, 219, 228}, {215, 219, 226}, {216, 219, 225}, {217, 220, 224}, {218, 220, 223}, {219, 220, 222}, {220, 220, 221}, {221, 220, 219}, {222, 219, 218}, {223, 219, 217}, {224, 218, 215}, {225, 218, 214}, {226, 217, 212}, {227, 217, 211}, {228, 216, 209}, {229, 216, 208}, {230, 215, 207}, {231, 214, 205}, {231, 214, 204}, {232, 213, 202}, {233, 212, 201}, {234, 211, 199}, {235, 211, 198}, {236, 210, 196}, {236, 209, 195}, {237, 208, 193}, {237, 207, 192}, {238, 207, 190}, {239, 206, 188}, {239, 205, 187}, {240, 204, 185}, {241, 203, 184}, {241, 202, 182}, {242, 201, 181}, {242, 200, 179}, {242, 199, 178}, {243, 198, 176}, {243, 197, 175}, {244, 196, 173}, {244, 195, 171}, {244, 194, 170}, {245, 193, 168}, {245, 192, 167}, {245, 191, 165}, {246, 189, 164}, {246, 188, 162}, {246, 187, 160}, {246, 186, 159}, {246, 185, 157}, {246, 183, 156}, {246, 182, 154}, {247, 181, 152}, {247, 179, 151}, {247, 178, 149}, {247, 177, 148}, {247, 176, 146}, {247, 174, 145}, {247, 173, 143}, {246, 171, 141}, {246, 170, 140}, {246, 169, 138}, {246, 167, 137}, {246, 166, 135}, {246, 164, 134}, {246, 163, 132}, {245, 161, 130}, {245, 160, 129}, {245, 158, 127}, {244, 157, 126}, {244, 155, 124}, {244, 154, 123}, {243, 152, 121}, {243, 150, 120}, {243, 149, 118}, {242, 147, 117}, {242, 145, 115}, {241, 144, 114}, {241, 142, 112}, {240, 141, 111}, {240, 139, 109}, {239, 137, 108}, {238, 135, 106}, {238, 134, 105}, {237, 132, 103}, {236, 130, 102}, {236, 128, 100}, {235, 127, 99}, {234, 125, 97}, {234, 123, 96}, {233, 121, 94}, {232, 119, 93}, {231, 117, 92}, {230, 116, 90}, {230, 114, 89}, {229, 112, 87}, {228, 110, 86}, {227, 108, 84}, {226, 106, 83}, {225, 104, 82}, {224, 102, 80}, {223, 100, 79}, {222, 98, 78}, {221, 96, 76}, {220, 94, 75}, {219, 92, 74}, {218, 90, 72}, {217, 88, 71}, {216, 86, 70}, {215, 84, 68}, {214, 82, 67}, {212, 79, 66}, {211, 77, 64}, {210, 75, 63}, {209, 73, 62}, {207, 70, 61}, {206, 68, 60}, {205, 66, 58}, {204, 63, 57}, {202, 61, 56}, {201, 59, 55}, {200, 56, 53}, {198, 53, 52}, {197, 50, 51}, {196, 48, 50}, {194, 45, 49}, {193, 42, 48}, {191, 40, 46}, {190, 35, 45}, {188, 31, 44}, {187, 26, 43}, {185, 22, 42}, {184, 17, 41}, {182, 13, 40}, {181, 8, 39}, {179, 3, 38},
	}

	for i, c := range colors {
		colormap[i] = color.NRGBA{
			R: uint8(c[0]),
			G: uint8(c[1]),
			B: uint8(c[2]),
			A: 255, // Opacity will be applied when generating image
		}
	}
}

// getColor maps a value in [0, 1] to a color using the colormap
// This matches matplotlib's colormap behavior exactly
func getColor(t float64, opacity float64) color.NRGBA {
	// Clamp to [0, 1]
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// Map to colormap index [0, 255]
	idx := t * 255.0
	idx1 := int(math.Floor(idx))
	idx2 := idx1 + 1

	// Handle edge cases
	if idx1 >= 255 {
		c := colormap[255]
		return color.NRGBA{
			R: c.R,
			G: c.G,
			B: c.B,
			A: uint8(math.Round(255 * opacity)),
		}
	}
	if idx1 < 0 {
		c := colormap[0]
		return color.NRGBA{
			R: c.R,
			G: c.G,
			B: c.B,
			A: uint8(math.Round(255 * opacity)),
		}
	}

	// Linear interpolation between colormap entries
	frac := idx - float64(idx1)
	c1 := colormap[idx1]
	c2 := colormap[idx2]

	return color.NRGBA{
		R: uint8(math.Round(float64(c1.R)*(1-frac) + float64(c2.R)*frac)),
		G: uint8(math.Round(float64(c1.G)*(1-frac) + float64(c2.G)*frac)),
		B: uint8(math.Round(float64(c1.B)*(1-frac) + float64(c2.B)*frac)),
		A: uint8(math.Round(255 * opacity)),
	}
}

// bilinearInterpolate performs bilinear interpolation matching matplotlib
func bilinearInterpolate(data [][]float64, x, y float64) float64 {
	rows := len(data)
	cols := len(data[0])

	// Clamp coordinates (matplotlib's behavior)
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= float64(cols-1) {
		x = float64(cols - 1)
	}
	if y >= float64(rows-1) {
		y = float64(rows - 1)
	}

	x1 := int(math.Floor(x))
	y1 := int(math.Floor(y))
	x2 := x1 + 1
	y2 := y1 + 1

	// Ensure bounds
	if x2 >= cols {
		x2 = cols - 1
	}
	if y2 >= rows {
		y2 = rows - 1
	}

	// Get corner values
	q11 := data[y1][x1]
	q12 := data[y2][x1]
	q21 := data[y1][x2]
	q22 := data[y2][x2]

	// Interpolation weights
	fx := x - float64(x1)
	fy := y - float64(y1)

	// Bilinear interpolation
	return (1-fx)*(1-fy)*q11 +
		fx*(1-fy)*q21 +
		(1-fx)*fy*q12 +
		fx*fy*q22
}

// createSpectrogram creates a spectrogram from a vector
func createSpectrogram(vector []float64, windowSize int) [][]float64 {
	nWindows := len(vector) / windowSize
	spectrogram := make([][]float64, windowSize)

	for i := 0; i < windowSize; i++ {
		spectrogram[i] = make([]float64, nWindows)
	}

	for i := 0; i < nWindows; i++ {
		start := i * windowSize
		for j := 0; j < windowSize; j++ {
			if start+j < len(vector) {
				spectrogram[j][i] = vector[start+j]
			}
		}
	}

	return spectrogram
}

// GenerateSpectrogramImage generates a spectrogram image from a 512-dimensional vector
// and writes it to the provided writer
func GenerateSpectrogramImage(vector []float64, width, height int, windowSize int, opacity float64, w io.Writer) error {
	// Create spectrogram
	spectrogram := createSpectrogram(vector, windowSize)

	// Find min/max for normalization (matplotlib does this automatically)
	minVal := math.Inf(1)
	maxVal := math.Inf(-1)

	for _, row := range spectrogram {
		for _, val := range row {
			if val < minVal {
				minVal = val
			}
			if val > maxVal {
				maxVal = val
			}
		}
	}

	rows := len(spectrogram)
	cols := len(spectrogram[0])

	// Matplotlib's imshow with extent=[0, n_windows, 0, window_size] means:
	// - x axis: image pixel [0, width] maps to data coordinate [0, cols]
	// - y axis: image pixel [0, height] maps to data coordinate [0, rows] (origin='lower' flips Y)
	// - aspect='auto' means pixels are scaled to fit the extent

	// Create image at target resolution
	// Use NRGBA (non-premultiplied alpha) which supports transparency
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Scale factors: map image coordinates to data coordinates
	// extent defines the data coordinate range that maps to the image
	dataWidth := float64(cols)   // n_windows
	dataHeight := float64(rows)  // window_size

	// For each pixel in the output image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map image pixel to data coordinates
			// extent=[0, n_windows, 0, window_size] with origin='lower'
			// means: x maps linearly, y is flipped
			
			// X: [0, width] -> [0, dataWidth]
			dataX := (float64(x) + 0.5) * dataWidth / float64(width)
			
			// Y: [0, height] -> [dataHeight, 0] (flipped for origin='lower')
			dataY := (float64(height-1-y) + 0.5) * dataHeight / float64(height)

			// Bilinear interpolation
			val := bilinearInterpolate(spectrogram, dataX, dataY)

			// Normalize to [0, 1] for colormap (matplotlib does this automatically)
			normalized := (val - minVal) / (maxVal - minVal)
			if maxVal == minVal {
				normalized = 0.5
			}

			// Get color from colormap
			c := getColor(normalized, opacity)

			// Set pixel
			img.Set(x, y, c)
		}
	}

	// Encode WebP with transparency (lossless)
	if err := nativewebp.Encode(w, img, nil); err != nil {
		return err
	}

	return nil
}

