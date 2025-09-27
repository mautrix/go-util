// Copyright (c) 2025 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package progver

import (
	"fmt"
	"runtime"
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

func (pv ProgramVersion) Init(tag, commit, rawBuildTime string) ProgramVersion {
	pv.Tag = tag
	baseVersion := strings.TrimPrefix(pv.BaseVersion, "v")
	tag = strings.TrimPrefix(tag, "v")
	if pv.SemCalVer && len(tag) > 0 {
		tag = semverToCalver(tag)
	}
	if tag == baseVersion {
		pv.IsRelease = true
		pv.LinkifiedVersion = fmt.Sprintf("[v%s](%s/releases/v%s)", baseVersion, pv.URL, pv.Tag)
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
	pv.BuildTime = buildTime
	return pv
}

func semverToCalver(semver string) string {
	parts := strings.SplitN(semver, ".", 3)
	if len(parts) < 2 {
		panic(fmt.Errorf("invalid semver for calendar versioning: %s", semver))
	}
	if len(parts[1]) != 4 {
		panic(fmt.Errorf("invalid minor semver component for calendar versioning: %s", parts[1]))
	}
	calver := parts[1][:2] + "." + parts[1][2:]
	if len(parts) == 3 {
		calver += "." + parts[2]
	}
	return calver
}
