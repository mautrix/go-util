// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog"

	"go.mau.fi/util/exzerolog"
)

type Format struct {
	Filename       string            `json:"filename"`
	NBStreams      int               `json:"nb_streams"`
	NBPrograms     int               `json:"nb_programs"`
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	StartTime      float64           `json:"start_time,string"`
	Duration       float64           `json:"duration,string"`
	Size           int               `json:"size,string"`
	BitRate        int               `json:"bit_rate,string"`
	ProbeScore     int               `json:"probe_score"`
	Tags           map[string]string `json:"tags"`
}

type Disposition struct {
	Default         int `json:"default"`
	Dub             int `json:"dub"`
	Original        int `json:"original"`
	Comment         int `json:"comment"`
	Lyrics          int `json:"lyrics"`
	Karaoke         int `json:"karaoke"`
	Forced          int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
	VisualImpaired  int `json:"visual_impaired"`
	CleanEffects    int `json:"clean_effects"`
	AttachedPic     int `json:"attached_pic"`
	TimedThumbnails int `json:"timed_thumbnails"`
	NonDiegetic     int `json:"non_diegetic"`
	Captions        int `json:"captions"`
	Descriptions    int `json:"descriptions"`
	Metadata        int `json:"metadata"`
	Dependent       int `json:"dependent"`
	StillImage      int `json:"still_image"`
}

type Stream struct {
	Index            int               `json:"index"`
	CodecName        string            `json:"codec_name"`
	CodecLongName    string            `json:"codec_long_name"`
	Profile          string            `json:"profile"`
	CodecType        string            `json:"codec_type"`
	CodecTagString   string            `json:"codec_tag_string"`
	CodecTag         string            `json:"codec_tag"`
	Width            int               `json:"width"`
	Height           int               `json:"height"`
	CodedWidth       int               `json:"coded_width"`
	CodedHeight      int               `json:"coded_height"`
	ClosedCaptions   int               `json:"closed_captions"`
	FilmGrain        int               `json:"film_grain"`
	HasBFrames       int               `json:"has_b_frames"`
	PixFmt           string            `json:"pix_fmt"`
	Level            int               `json:"level"`
	ColorRange       string            `json:"color_range"`
	ColorSpace       string            `json:"color_space"`
	ColorTransfer    string            `json:"color_transfer"`
	ColorPrimaries   string            `json:"color_primaries"`
	ChromaLocation   string            `json:"chroma_location"`
	FieldOrder       string            `json:"field_order"`
	Refs             int               `json:"refs"`
	IsAvc            string            `json:"is_avc"`
	NalLengthSize    string            `json:"nal_length_size"`
	ID               string            `json:"id"`
	RFrameRate       string            `json:"r_frame_rate"`
	AvgFrameRate     string            `json:"avg_frame_rate"`
	TimeBase         string            `json:"time_base"`
	StartPts         int               `json:"start_pts"`
	StartTime        float64           `json:"start_time,string"`
	DurationTS       int               `json:"duration_ts"`
	Duration         float64           `json:"duration,string"`
	BitRate          int               `json:"bit_rate,string"`
	BitsPerRawSample int               `json:"bits_per_raw_sample,string"`
	NumberOfFrames   int               `json:"nb_frames,string"`
	ExtradataSize    int               `json:"extradata_size"`
	Disposition      Disposition       `json:"disposition"`
	Tags             map[string]string `json:"tags"`
	SampleFormat     string            `json:"sample_fmt"`
	SampleRate       int               `json:"sample_rate,string"`
	Channels         int               `json:"channels"`
	ChannelLayout    string            `json:"channel_layout"`
	BitsPerSample    int               `json:"bits_per_sample"`
	InitialPadding   int               `json:"initial_padding"`
}

type ProbeResult struct {
	Streams []*Stream `json:"streams"`
	Format  *Format   `json:"format"`
}

var ffprobeDefaultParams = []string{"-hide_banner", "-loglevel", "warning", "-print_format", "json", "-show_format", "-show_streams"}

var probeExecutor = func(ctx context.Context, args []string) ([]byte, error) {
	if ffprobePath == "" {
		return nil, fmt.Errorf("ffprobe not found in $PATH")
	}
	ctxLog := zerolog.Ctx(ctx).With().Str("command", "ffmpeg").Logger()
	logWriter := exzerolog.NewLogWriter(ctxLog).WithLevel(zerolog.WarnLevel)
	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	var stdout bytes.Buffer
	cmd.Stderr = logWriter
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.Bytes(), err
}

// SetProbeExecutor overrides the function used to create the ffprobe subprocess.
//
// The function gets the full list of command-line arguments to pass to ffprobe,
// and is expected to return the stdout of the command, or an error if the command returns a non-zero exit code.
//
// Note that [ProbeSupported] will still return false if ffprobe wasn't found in $PATH and [SetProbePath] wasn't called.
func SetProbeExecutor(fn func(context.Context, []string) ([]byte, error)) {
	probeExecutor = fn
}

func Probe(ctx context.Context, path string) (*ProbeResult, error) {
	stdout, err := probeExecutor(ctx, append(ffprobeDefaultParams, path))
	if err != nil {
		return nil, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var result ProbeResult
	err = json.Unmarshal(stdout, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ffmpeg output: %w", err)
	}

	return &result, nil
}
