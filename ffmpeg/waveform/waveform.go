// Copyright (c) 2025 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package waveform

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"slices"

	"go.mau.fi/util/ffmpeg"
)

func makeWaveformArgs(samples, maxValue int) []string {
	return []string{
		"-filter_complex",
		fmt.Sprintf("aformat=channel_layouts=mono,showwavespic=s=%dx%d:colors=white", samples, maxValue*2),
		"-frames:v",
		"1",
		"-update",
		"1",
	}
}

func Generate(ctx context.Context, inputFile string, samples, maxValue int) ([]int, error) {
	tempFile, err := os.CreateTemp("", "waveform-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	_ = tempFile.Close()
	_ = os.Remove(tempFile.Name())
	err = ffmpeg.ConvertPathWithDestination(ctx, inputFile, tempFile.Name(), nil, makeWaveformArgs(samples, maxValue), false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate waveform png with ffmpeg: %w", err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()
	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to reopen temp file: %w", err)
	}
	defer func() {
		_ = tempFile.Close()
	}()
	decoded, err := png.Decode(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode waveform png: %w", err)
	}
	return parseWaveformImage(decoded, maxValue), nil
}

func GenerateBytes(ctx context.Context, inputData []byte, inputMime string, samples, maxValue int) ([]int, error) {
	waveformBytes, err := ffmpeg.ConvertBytes(ctx, inputData, ".png", nil, makeWaveformArgs(samples, maxValue), inputMime)
	if err != nil {
		return nil, fmt.Errorf("failed to generate waveform png with ffmpeg: %w", err)
	}
	decoded, err := png.Decode(bytes.NewReader(waveformBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode waveform png: %w", err)
	}
	return parseWaveformImage(decoded, maxValue), nil
}

func isWhite(c color.Color) bool {
	r, g, b, a := c.RGBA()
	return a > 0 && r > 0x7FFF && g > 0x7FFF && b > 0x7FFF
}

func findAvgMinMax(img image.Image, bounds image.Rectangle, x, targetMaxVal int) int {
	var topVal, bottomVal int
	for topVal = targetMaxVal; topVal > 0; topVal-- {
		if isWhite(img.At(bounds.Min.X+x, bounds.Min.Y+targetMaxVal-topVal)) {
			break
		}
	}
	for bottomVal = targetMaxVal; bottomVal > 0; bottomVal-- {
		if isWhite(img.At(bounds.Min.X+x, bounds.Max.Y-targetMaxVal+bottomVal)) {
			break
		}
	}
	return (topVal + bottomVal) / 2
}

func clamp(data []int, to int) {
	maxVal := slices.Max(data)
	if maxVal < to {
		for i := range data {
			data[i] = int(float64(data[i]) * float64(to) / float64(maxVal))
		}
	}
}

func parseWaveformImage(img image.Image, targetMaxVal int) []int {
	bounds := img.Bounds()
	out := make([]int, bounds.Dx())
	for x := 0; x < len(out); x++ {
		out[x] = findAvgMinMax(img, bounds, x, targetMaxVal)
	}
	clamp(out, targetMaxVal)
	return out
}
