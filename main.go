package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// Config defines the LUT parameters.
type Config struct {
	Size           int     `json:"size"`            // Grid dimension (default 17)
	RedTint        float64 `json:"red_tint"`        // Additional red multiplier (if used in creative look)
	BlueTint       float64 `json:"blue_tint"`       // Additional blue multiplier (if used in creative look)
	Output         string  `json:"output"`          // Output file name (e.g., "apple_log_cinematic.cube")
	Look           string  `json:"look"`            // "none", "tealOrange", or "warmVintage"
	ExposureOffset float64 `json:"exposure_offset"` // Factor to adjust exposure (default 1.0)
}

func (c *Config) setDefaults() {
	if c.Size <= 0 {
		c.Size = 17
	}
	if c.RedTint == 0 {
		c.RedTint = 1.05
	}
	if c.BlueTint == 0 {
		c.BlueTint = 0.95
	}
	if c.Output == "" {
		c.Output = "output.cube"
	}
	if c.Look == "" {
		c.Look = "none"
	}
	if c.ExposureOffset == 0 {
		c.ExposureOffset = 1.0
	}
}

// appleLogToLinear approximates the decoding of Apple Log to linear light.
// This is a simplified function; in practice, use the official curve.
func appleLogToLinear(x float64, exposureOffset float64) float64 {
	// Apply an exposure offset and clip to [0,1]
	v := min(x*exposureOffset, 1)
	// A simple power function to approximate the inverse log curve.
	// (Note: This is a rough approximation.)
	return math.Pow(v, 1.5)
}

// rec2020ToRec709 converts Rec.2020 linear values to Rec.709 linear using a 3x3 matrix.
func rec2020ToRec709(r, g, b float64) (float64, float64, float64) {
	// Matrix coefficients (approximation)
	r709 := 1.660*r - 0.587*g - 0.073*b
	g709 := -0.124*r + 1.132*g - 0.008*b
	b709 := -0.018*r - 0.100*g + 1.118*b
	// Clip values to [0,1]
	if r709 < 0 {
		r709 = 0
	}
	if g709 < 0 {
		g709 = 0
	}
	if b709 < 0 {
		b709 = 0
	}
	if r709 > 1 {
		r709 = 1
	}
	if g709 > 1 {
		g709 = 1
	}
	if b709 > 1 {
		b709 = 1
	}
	return r709, g709, b709
}

// rec709OETF applies the Rec.709 opto-electronic transfer function.
func rec709OETF(linear float64) float64 {
	if linear < 0.018 {
		return 4.5 * linear
	}
	return 1.099*math.Pow(linear, 0.45) - 0.099
}

// applyTealOrange applies a simplified teal & orange look.
func applyTealOrange(r, g, b float64) (float64, float64, float64) {
	// Compute luminance
	lum := 0.2126*r + 0.7152*g + 0.0722*b
	origR, origG, origB := r, g, b
	if lum < 0.5 {
		// In shadows, reduce red slightly and boost blue
		rNew := r * 0.95
		bNew := b * 1.1
		// Blend the original with the modified values
		r = 0.7*origR + 0.3*rNew
		g = 0.7*origG + 0.3*origG // green remains similar
		b = 0.7*origB + 0.3*bNew
	} else {
		// In highlights, boost red and reduce blue
		rNew := r * 1.1
		bNew := b * 0.95
		r = 0.7*origR + 0.3*rNew
		g = 0.7*origG + 0.3*origG
		b = 0.7*origB + 0.3*bNew
	}
	if r > 1 {
		r = 1
	}
	if g > 1 {
		g = 1
	}
	if b > 1 {
		b = 1
	}
	return r, g, b
}

// applyWarmVintage applies a simplified warm vintage look.
func applyWarmVintage(r, g, b float64) (float64, float64, float64) {
	// Apply a subtle warm tint: increase red slightly, decrease blue
	r = r * 1.05
	b = b * 0.95
	// Optionally, lower contrast gently by blending with mid-gray (0.5)
	r = 0.9*r + 0.1*0.5
	g = 0.9*g + 0.1*0.5
	b = 0.9*b + 0.1*0.5
	if r > 1 {
		r = 1
	}
	if g > 1 {
		g = 1
	}
	if b > 1 {
		b = 1
	}
	return r, g, b
}

// generateLUT creates the LUT as a string based on the config.
// For each input grid value (representing an Apple Log encoded value), we:
// 1. Decode from Apple Log to linear light.
// 2. Convert from Rec.2020 (linear) to Rec.709 (linear).
// 3. Apply Rec.709 OETF (gamma encoding).
// 4. Optionally, apply a creative look.
func generateLUT(cfg Config) string {
	size := cfg.Size
	var builder strings.Builder

	// Write LUT header
	builder.WriteString("# Generated Cinematic LUT for Apple Log to Rec.709 conversion\n")
	builder.WriteString(fmt.Sprintf("LUT_3D_SIZE %d\n", size))

	// Loop over the 3D LUT grid.
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			for k := 0; k < size; k++ {
				// Normalized input values (simulate Apple Log encoded values).
				// These are in the range [0, 1].
				inR := float64(i) / float64(size-1)
				inG := float64(j) / float64(size-1)
				inB := float64(k) / float64(size-1)

				// Step 1: Decode Apple Log to linear light.
				linR := appleLogToLinear(inR, cfg.ExposureOffset)
				linG := appleLogToLinear(inG, cfg.ExposureOffset)
				linB := appleLogToLinear(inB, cfg.ExposureOffset)

				// Step 2: Convert from Rec.2020 (linear) to Rec.709 (linear).
				convR, convG, convB := rec2020ToRec709(linR, linG, linB)

				// Step 3: Encode using Rec.709 OETF.
				encR := rec709OETF(convR)
				encG := rec709OETF(convG)
				encB := rec709OETF(convB)

				// Step 4: Apply creative look if specified.
				switch strings.ToLower(cfg.Look) {
				case "tealorange":
					encR, encG, encB = applyTealOrange(encR, encG, encB)
				case "warmvintage":
					encR, encG, encB = applyWarmVintage(encR, encG, encB)
				}

				// Write the LUT line with 6 decimal places.
				builder.WriteString(fmt.Sprintf("%.6f %.6f %.6f\n", encR, encG, encB))
			}
		}
	}
	return builder.String()
}

// processConfigFile reads a config JSON file, generates LUT data, and writes the .cube file.
func processConfigFile(configPath, outputDir string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Error reading config file %s: %v\n", configPath, err)
		return
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("Error parsing JSON in %s: %v\n", configPath, err)
		return
	}
	cfg.setDefaults()

	lutData := generateLUT(cfg)

	// Determine the output file name.
	outFileName := cfg.Output
	// If not an absolute path, use the output directory.
	if !filepath.IsAbs(outFileName) {
		outFileName = filepath.Join(outputDir, outFileName)
	}

	if err := os.WriteFile(outFileName, []byte(lutData), 0644); err != nil {
		log.Printf("Error writing output file %s: %v\n", outFileName, err)
		return
	}
	log.Printf("LUT successfully written to %s\n", outFileName)
}

func main() {
	// Command-line flags for directories.
	configDir := flag.String("configDir", "configs", "Directory containing JSON config files")
	outputDir := flag.String("outputDir", "output", "Directory to write the generated .cube files")
	flag.Parse()

	// Ensure output directory exists.
	if err := os.MkdirAll(*outputDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Walk through the config directory and process each JSON file.
	err := filepath.Walk(*configDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			log.Printf("Processing config: %s\n", path)
			processConfigFile(path, *outputDir)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking through config directory: %v", err)
	}
}
