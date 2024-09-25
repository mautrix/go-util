// Copyright (c) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package lottie

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"

	"go.mau.fi/util/exzerolog"
	"go.mau.fi/util/ffmpeg"
)

var lottieconverterPath string

func init() {
	lottieconverterPath, _ = exec.LookPath("lottieconverter")
}

// Supported returns whether lottieconverter is available on the system.
//
// lottieconverter is considered to be available if a binary called
// lottieconverter is found in $PATH, or if [SetPath] has been called
// explicitly with a non-empty path.
func Supported() bool {
	return lottieconverterPath != ""
}

// SetPath overrides the path to the lottieconverter binary.
func SetPath(path string) {
	lottieconverterPath = path
}

// Convert converts lottie data an image or image(s) using lottieconverter.
//
// Args:
//   - input: an io.Reader containing the lottie data to convert.
//   - outputFilename: the filename to write the output to.
//   - outputWriter: an io.Writer to write the output to.
//   - format: the output format. Can be one of: png, gif, or pngs.
//   - width: the width of the output image(s).
//   - height: the height of the output image(s).
//   - extraArgs: additional arguments to pass to lottieconverter.
//
// The outputFilename and outputWriter parameters are mutually exclusive.
func Convert(ctx context.Context, input io.Reader, outputFilename string, outputWriter io.Writer, format string, width, height int, extraArgs ...string) error {
	// Verify the input parameters and calculate the actual outputFilename that
	// will be used when shelling out to lottieconverter.
	//
	// We are panicking here because it's a programming error to call this
	// function with invalid parameters.
	if outputFilename == "" && outputWriter == nil {
		panic("lottie.Convert: either outputFile or outputWriter must be provided")
	} else if outputWriter != nil {
		if outputFilename != "" {
			panic("lottie.Convert: only one of outputFile or outputWriter can be provided")
		}
		outputFilename = "-"
	}

	args := []string{"-", outputFilename, format, fmt.Sprintf("%dx%d", width, height)}
	args = append(args, extraArgs...)

	cmd := exec.CommandContext(ctx, lottieconverterPath, args...)
	cmd.Stdin = input
	cmd.Stdout = outputWriter
	log := zerolog.Ctx(ctx).With().Str("command", "lottieconverter").Logger()
	cmd.Stderr = exzerolog.NewLogWriter(log).WithLevel(zerolog.WarnLevel)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("lottieconverter error: %w", err)
	}
	return nil
}

// FFmpegConvert converts lottie data to a video or image using ffmpeg.
//
// This function should only be called if [ffmpeg.Supported] returns true.
//
// Args:
//   - input: an io.Reader containing the lottie data to convert.
//   - outputFile: the filename to write the output to. Must have .webp or .webm extension.
//   - width: the width of the output video or image.
//   - height: the height of the output video or image.
//   - fps: the framerate of the output video.
//
// Returns: the converted data as a *bytes.Buffer, the mimetype of the output,
// and the thumbnail data as a PNG.
func FFmpegConvert(ctx context.Context, input io.Reader, outputFile string, width, height, fps int) (thumbnailData []byte, err error) {
	if !ffmpeg.Supported() {
		return nil, fmt.Errorf("ffmpeg is not available")
	}

	tmpDir, err := os.MkdirTemp("", "lottieconvert")
	if err != nil {
		return
	}
	defer os.RemoveAll(tmpDir)

	err = Convert(ctx, input, tmpDir+"/out_", nil, "pngs", width, height, strconv.Itoa(fps))
	if err != nil {
		return
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return
	}

	var firstFrameName string
	for _, file := range files {
		if firstFrameName == "" || file.Name() < firstFrameName {
			firstFrameName = file.Name()
		}
	}
	thumbnailData, err = os.ReadFile(fmt.Sprintf("%s/%s", tmpDir, firstFrameName))
	if err != nil {
		return
	}

	var outputArgs []string
	switch filepath.Ext(outputFile) {
	case ".webm":
		outputArgs = []string{"-c:v", "libvpx-vp9", "-pix_fmt", "yuva420p", "-f", "webm"}
	case ".webp":
		outputArgs = []string{"-c:v", "libwebp_anim", "-pix_fmt", "yuva420p", "-f", "webp"}
	default:
		err = fmt.Errorf("unsupported extension %s", filepath.Ext(outputFile))
		return
	}
	err = ffmpeg.ConvertPathWithDestination(
		ctx,
		tmpDir+"/out_*.png",
		outputFile,
		[]string{"-framerate", strconv.Itoa(fps), "-pattern_type", "glob"},
		outputArgs,
		false,
	)
	return
}
