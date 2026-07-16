// Copyright (c) 2026 Archmagece
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

// GetTimezone resolves the current timezone via /etc/localtime on macOS.
// Canceling ctx aborts the underlying platform query.
func GetTimezone(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "readlink", "/etc/localtime").Output()
	if err != nil {
		return "", err
	}
	return ParseTimezoneLink(string(out)), nil
}
