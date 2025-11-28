// Copyright (c) 2025 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package progver

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type ProgramVersion struct {
	// These should be hardcoded
	Name        string
	URL         string
	BaseVersion string
	SemCalVer   bool

	// These are set by the values passed to InitVersion
	Commit    string
	Tag       string
	BuildTime time.Time

	// These are computed by InitVersion
	IsRelease          bool
	FormattedVersion   string
	LinkifiedVersion   string
	VersionDescription string
}

func (pv ProgramVersion) MarkdownDescription() string {
	return fmt.Sprintf("[%s](%s) %s (%s)", pv.Name, pv.URL, pv.LinkifiedVersion, pv.BuildTime.Format(time.RFC1123))
}

func findCommitFromBuildInfo() string {
	info, _ := debug.ReadBuildInfo()
	if info == nil {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && len(setting.Value) >= 40 {
			return setting.Value
		}
	}
	return ""
}

func (pv ProgramVersion) Init(tag, commit, rawBuildTime string) ProgramVersion {
	if commit == "" || commit == "unknown" {
		commit = findCommitFromBuildInfo()
	}
	if tag == "unknown" {
		tag = ""
	}
	pv.Tag = tag
	baseVersion := strings.TrimPrefix(pv.BaseVersion, "v")
	tag = strings.TrimPrefix(tag, "v")
	if pv.SemCalVer && len(tag) > 0 {
		tag = semverToCalver(tag)
	}
	if tag == baseVersion || tag == baseVersion+".0" {
		pv.IsRelease = true
		pv.FormattedVersion = "v" + baseVersion
		pv.LinkifiedVersion = fmt.Sprintf("[%s](%s/releases/%s)", pv.FormattedVersion, pv.URL, pv.Tag)
	} else {
		suffix := ""
		if !strings.HasSuffix(baseVersion, "+dev") {
			suffix = "+dev"
		}
		if len(commit) > 8 {
			pv.FormattedVersion = fmt.Sprintf("v%s%s.%s", baseVersion, suffix, commit[:8])
		} else {
			pv.FormattedVersion = fmt.Sprintf("v%s%s.unknown", baseVersion, suffix)
		}
		if len(commit) > 8 {
			pv.LinkifiedVersion = strings.Replace(pv.FormattedVersion, commit[:8], fmt.Sprintf("[%s](%s/commit/%s)", commit[:8], pv.URL, commit), 1)
		} else {
			pv.LinkifiedVersion = pv.FormattedVersion
		}
	}
	var buildTime time.Time
	if rawBuildTime != "unknown" {
		buildTime, _ = time.Parse(time.RFC3339, rawBuildTime)
	}
	var builtWith string
	if buildTime.IsZero() {
		rawBuildTime = "unknown"
		builtWith = fmt.Sprintf("built with %s", runtime.Version())
	} else {
		rawBuildTime = buildTime.Format(time.RFC1123)
		builtWith = fmt.Sprintf("built at %s with %s", rawBuildTime, runtime.Version())
	}
	pv.VersionDescription = fmt.Sprintf("%s %s (%s)", pv.Name, pv.FormattedVersion, builtWith)
	pv.Commit = commit
	if pv.Commit == "" {
		pv.Commit = "unknown"
	}
	pv.BuildTime = buildTime
	return pv
}

func semverToCalver(semver string) string {
	parts := strings.SplitN(semver, ".", 3)
	if len(parts) != 3 {
		panic(fmt.Errorf("invalid semver: %s", semver))
	}
	if len(parts[1]) != 4 {
		panic(fmt.Errorf("invalid minor semver component for calendar versioning: %s", parts[1]))
	}
	return parts[1][:2] + "." + parts[1][2:] + "." + parts[2]
}
