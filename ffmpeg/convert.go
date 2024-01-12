// Copyright (c) 2022 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"go.mau.fi/util/exmime"
	"go.mau.fi/util/exzerolog"
)

var ffmpegDefaultParams = []string{"-hide_banner", "-loglevel", "warning"}

var ffmpegPath string

func init() {
	ffmpegPath, _ = exec.LookPath("ffmpeg")
}

// Supported returns whether ffmpeg is available on the system.
//
// ffmpeg is considered to be available if a binary called ffmpeg is found in $PATH,
// or if [SetPath] has been called explicitly with a non-empty path.
func Supported() bool {
	return ffmpegPath != ""
}

// SetPath overrides the path to the ffmpeg binary.
func SetPath(path string) {
	ffmpegPath = path
}

// ConvertPath converts a media file on the disk using ffmpeg.
//
// Args:
// * inputFile: The full path to the file.
// * outputExtension: The extension that the output file should be.
// * inputArgs: Arguments to tell ffmpeg how to parse the input file.
// * outputArgs: Arguments to tell ffmpeg how to convert the file to reach the wanted output.
// * removeInput: Whether the input file should be removed after converting.
//
// Returns: the path to the converted file.
func ConvertPath(ctx context.Context, inputFile string, outputExtension string, inputArgs []string, outputArgs []string, removeInput bool) (string, error) {
	outputFilename := strings.TrimSuffix(strings.TrimSuffix(inputFile, filepath.Ext(inputFile)), "*") + outputExtension

	args := make([]string, 0, len(ffmpegDefaultParams)+len(inputArgs)+2+len(outputArgs)+1)
	args = append(args, ffmpegDefaultParams...)
	args = append(args, inputArgs...)
	args = append(args, "-i", inputFile)
	args = append(args, outputArgs...)
	args = append(args, outputFilename)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	ctxLog := zerolog.Ctx(ctx).With().Str("command", "ffmpeg").Logger()
	logWriter := exzerolog.NewLogWriter(ctxLog).WithLevel(zerolog.WarnLevel)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %+v", err)
	}

	if removeInput {
		_ = os.Remove(inputFile)
	}

	return outputFilename, nil
}

// ConvertBytes converts media data using ffmpeg.
//
// Args:
// * data: The media data to convert
// * outputExtension: The extension that the output file should be.
// * inputArgs: Arguments to tell ffmpeg how to parse the input file.
// * outputArgs: Arguments to tell ffmpeg how to convert the file to reach the wanted output.
// * inputMime: The mimetype of the input data.
//
// Returns: the converted data
func ConvertBytes(ctx context.Context, data []byte, outputExtension string, inputArgs []string, outputArgs []string, inputMime string) ([]byte, error) {
	tempdir, err := os.MkdirTemp("", "mautrix_ffmpeg_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempdir)
	inputFileName := fmt.Sprintf("%s/input%s", tempdir, exmime.ExtensionFromMimetype(inputMime))

	inputFile, err := os.OpenFile(inputFileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	_, err = inputFile.Write(data)
	if err != nil {
		_ = inputFile.Close()
		return nil, fmt.Errorf("failed to write data to input file: %w", err)
	}
	_ = inputFile.Close()

	outputPath, err := ConvertPath(ctx, inputFileName, outputExtension, inputArgs, outputArgs, false)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(outputPath)
}
