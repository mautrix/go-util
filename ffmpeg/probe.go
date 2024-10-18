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

func Probe(ctx context.Context, path string) (*ProbeResult, error) {
	ctxLog := zerolog.Ctx(ctx).With().Str("command", "ffmpeg").Logger()
	logWriter := exzerolog.NewLogWriter(ctxLog).WithLevel(zerolog.WarnLevel)
	cmd := exec.CommandContext(ctx, ffprobePath, append(ffprobeDefaultParams, path)...)
	var stdout bytes.Buffer
	cmd.Stderr = logWriter
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var result ProbeResult
	err = json.NewDecoder(&stdout).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ffmpeg output: %w", err)
	}

	return &result, nil
}
