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
	"os/exec"

	"github.com/rs/zerolog"
	"go.mau.fi/util/exzerolog"
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
