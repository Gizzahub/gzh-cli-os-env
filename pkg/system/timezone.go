// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"os/exec"
	"strings"
)

const zoneinfoPrefix = "zoneinfo/"

// ParseTimezoneLink extracts the IANA timezone from a /etc/localtime
// symlink target produced by `readlink /etc/localtime`, e.g.
//
//	/var/db/timezone/zoneinfo/Asia/Seoul -> "Asia/Seoul"
//	/usr/share/zoneinfo/Asia/Seoul       -> "Asia/Seoul"
//
// If the target does not contain "zoneinfo/", the basename is returned.
// Pure function so it can be tested without a real symlink.
func ParseTimezoneLink(link string) string {
	link = strings.TrimSpace(link)
	if idx := strings.Index(link, zoneinfoPrefix); idx >= 0 {
		return link[idx+len(zoneinfoPrefix):]
	}
	// Fall back to the basename for non-standard targets.
	if i := strings.LastIndex(link, "/"); i >= 0 {
		return link[i+1:]
	}
	return link
}

// ParseTimedatectlShow extracts Timezone from `timedatectl show` output.
// Looks for a line like:
//
//	Timezone=Asia/Seoul
//
// Pure function for offline unit tests.
func ParseTimedatectlShow(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Timezone=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Timezone="))
		}
	}
	return ""
}

// GetTimezone resolves the current timezone via /etc/localtime, with a
// timedatectl fallback on Linux when the symlink is unavailable.
// Canceling ctx aborts the underlying platform query.
func GetTimezone(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "readlink", "/etc/localtime").Output()
	if err == nil {
		return ParseTimezoneLink(string(out)), nil
	}
	// Linux fallback when /etc/localtime is not a readable symlink.
	td, err2 := exec.CommandContext(ctx, "timedatectl", "show").Output()
	if err2 != nil {
		return "", err
	}
	if tz := ParseTimedatectlShow(string(td)); tz != "" {
		return tz, nil
	}
	return "", err
}
