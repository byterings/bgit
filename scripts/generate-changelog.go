//go:build ignore
// +build ignore

// This script generates CHANGELOG.md from changelog.json
// Run with: go run scripts/generate-changelog.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Feature struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Command     *string `json:"command"`
	DocsLink    string  `json:"docs_link"`
}

type Release struct {
	Version         string    `json:"version"`
	Date            string    `json:"date"`
	Phase           int       `json:"phase"`
	Summary         string    `json:"summary"`
	Features        []Feature `json:"features"`
	Fixes           []Feature `json:"fixes"`
	BreakingChanges []Feature `json:"breaking_changes"`
}

type Changelog struct {
	Releases []Release `json:"releases"`
}

func main() {
	// Read changelog.json
	data, err := os.ReadFile("changelog.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading changelog.json: %v\n", err)
		os.Exit(1)
	}

	var changelog Changelog
	if err := json.Unmarshal(data, &changelog); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing changelog.json: %v\n", err)
		os.Exit(1)
	}

	// Generate markdown
	var sb strings.Builder

	sb.WriteString("# Changelog\n\n")
	sb.WriteString("All notable changes to bgit will be documented in this file.\n\n")
	sb.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n")
	sb.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")
	sb.WriteString("## [Unreleased]\n\n")

	for _, release := range changelog.Releases {
		// Version header
		dateStr := release.Date
		if dateStr == "" {
			dateStr = "TBD"
		}
		sb.WriteString(fmt.Sprintf("## [%s] - %s (Phase %d)\n\n", release.Version, dateStr, release.Phase))

		// Summary
		sb.WriteString(release.Summary + "\n\n")

		// Features
		if len(release.Features) > 0 {
			sb.WriteString("### Added\n")
			for _, feature := range release.Features {
				cmdStr := ""
				if feature.Command != nil && *feature.Command != "" {
					cmdStr = fmt.Sprintf(" (`%s`)", *feature.Command)
				}
				sb.WriteString(fmt.Sprintf("- **%s** - %s%s\n", feature.Title, feature.Description, cmdStr))
			}
			sb.WriteString("\n")
		}

		// Fixes
		if len(release.Fixes) > 0 {
			sb.WriteString("### Fixed\n")
			for _, fix := range release.Fixes {
				sb.WriteString(fmt.Sprintf("- **%s** - %s\n", fix.Title, fix.Description))
			}
			sb.WriteString("\n")
		}

		// Breaking changes
		if len(release.BreakingChanges) > 0 {
			sb.WriteString("### Breaking Changes\n")
			for _, change := range release.BreakingChanges {
				sb.WriteString(fmt.Sprintf("- **%s** - %s\n", change.Title, change.Description))
			}
			sb.WriteString("\n")
		}
	}

	// Write CHANGELOG.md
	if err := os.WriteFile("CHANGELOG.md", []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CHANGELOG.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("CHANGELOG.md generated successfully")
}
