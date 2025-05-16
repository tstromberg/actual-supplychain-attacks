package main

import (
	"encoding/csv"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Headers for CSV output
var csvHeaders = []string{
	"name", "title", "repo", "synopsis", "year", "start_date", "end_date",
	"attribution_type", "component_type", "lang", "cause", "motive",
	"transitive", "insertion_phase", "impact_type", "impact_user_count",
	"references", "versions", "commits", "historical_artifacts", "current_artifacts",
	"domain", "domain_type", "artifact_type", "hashes",
}

// AttackEntry defines the structure for unmarshalling the meta.yaml files
type AttackEntry struct {
	Name                string   `yaml:"name"`
	Title               string   `yaml:"title"`
	Repo                string   `yaml:"repo"`
	Synopsis            string   `yaml:"synopsis"`
	StartDate           string   `yaml:"start_date"`
	EndDate             string   `yaml:"end_date"`
	AttributionType     string   `yaml:"attribution_type"`
	ComponentType       string   `yaml:"component_type"`
	Lang                string   `yaml:"lang"`
	Cause               string   `yaml:"cause"`
	Motive              string   `yaml:"motive"`
	Transitive          bool     `yaml:"transitive"`
	InsertionPhase      string   `yaml:"insertion_phase"`
	ImpactType          string   `yaml:"impact_type"`
	ImpactUserCount     int      `yaml:"impact_user_count"`
	References          []string `yaml:"references"`
	Versions            []string `yaml:"versions"`
	Commits             []string `yaml:"commits"`
	HistoricalArtifacts []string `yaml:"historical_artifacts"`
	CurrentArtifacts    []string `yaml:"current_artifacts"`
	Domain              string   `yaml:"domain"`
	DomainType          string   `yaml:"domain_type"`
	ArtifactType        string   `yaml:"artifact_type"`
	Hashes              []string `yaml:"hashes"`
}

func main() {
	// Get search directory from args or use current directory
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalf("Error: Could not determine absolute path for %q: %v", rootDir, err)
	}

	log.Printf("Info: Searching for meta.yaml files in: %s", absRootDir)

	// Collect all entries from all meta.yaml files
	var allEntries []AttackEntry
	count := 0

	err = filepath.WalkDir(absRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "meta.yaml" {
			return nil
		}

		log.Printf("Info: Processing file: %s", path)

		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: Could not read %s: %v. Skipping.", path, err)
			return nil
		}

		var entries []AttackEntry
		if err := yaml.Unmarshal(data, &entries); err != nil {
			log.Printf("Warning: Could not parse %s: %v. Skipping.", path, err)
			return nil
		}

		allEntries = append(allEntries, entries...)
		count += len(entries)
		return nil
	})
	if err != nil {
		log.Fatalf("Error: Failed walking directory: %v", err)
	}

	if count == 0 {
		log.Println("Info: No meta.yaml files found or successfully processed.")
		return
	}

	// Sort entries by start_date
	sort.Slice(allEntries, func(i, j int) bool {
		// Empty dates go last
		if allEntries[i].StartDate == "" {
			return false
		}
		if allEntries[j].StartDate == "" {
			return true
		}

		// Try parsing as dates first
		date1, err1 := time.Parse("2006-01-02", allEntries[i].StartDate)
		date2, err2 := time.Parse("2006-01-02", allEntries[j].StartDate)
		if err1 == nil && err2 == nil {
			return date1.Before(date2)
		}

		// Fall back to string comparison
		return allEntries[i].StartDate < allEntries[j].StartDate
	})

	// Write sorted entries to CSV
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if err := writer.Write(csvHeaders); err != nil {
		log.Fatalf("Error: Failed to write CSV headers: %v", err)
	}

	for _, entry := range allEntries {
		row := entryToRow(entry)
		if err := writer.Write(row); err != nil {
			log.Printf("Warning: Error writing row: %v", err)
		}
	}

	log.Printf("Info: CSV generation complete. Processed %d entries.", count)
}

// entryToRow converts an AttackEntry to a CSV row
func entryToRow(entry AttackEntry) []string {
	row := make([]string, len(csvHeaders))

	for i, header := range csvHeaders {
		switch header {
		case "name":
			row[i] = entry.Name
		case "title":
			row[i] = entry.Title
		case "repo":
			row[i] = entry.Repo
		case "synopsis":
			row[i] = entry.Synopsis
		case "year":
			if len(entry.StartDate) >= 4 {
				row[i] = entry.StartDate[0:4]
			}
		case "start_date":
			row[i] = entry.StartDate
		case "end_date":
			row[i] = entry.EndDate
		case "attribution_type":
			row[i] = entry.AttributionType
		case "component_type":
			row[i] = entry.ComponentType
		case "lang":
			row[i] = entry.Lang
		case "cause":
			row[i] = entry.Cause
		case "motive":
			row[i] = entry.Motive
		case "transitive":
			row[i] = strconv.FormatBool(entry.Transitive)
		case "insertion_phase":
			row[i] = entry.InsertionPhase
		case "impact_type":
			row[i] = entry.ImpactType
		case "impact_user_count":
			if entry.ImpactUserCount != 0 {
				row[i] = strconv.Itoa(entry.ImpactUserCount)
			}
		case "references":
			row[i] = strings.Join(entry.References, ", ")
		case "versions":
			row[i] = strings.Join(entry.Versions, ", ")
		case "commits":
			row[i] = strings.Join(entry.Commits, ", ")
		case "historical_artifacts":
			row[i] = strings.Join(entry.HistoricalArtifacts, ", ")
		case "current_artifacts":
			row[i] = strings.Join(entry.CurrentArtifacts, ", ")
		case "domain":
			row[i] = entry.Domain
		case "domain_type":
			row[i] = entry.DomainType
		case "artifact_type":
			row[i] = entry.ArtifactType
		case "hashes":
			row[i] = strings.Join(entry.Hashes, ", ")
		}
	}

	return row
}
