package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	// Set up logging
	log.SetPrefix("")
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	// Get search directory
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalf("Error: Could not determine absolute path for %q: %v", rootDir, err)
	}

	log.Printf("Info: Searching for meta.yaml files in: %s", absRootDir)

	// Process YAML files and convert to CSV
	entries, err := processYAMLFiles(absRootDir)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(entries) == 0 {
		log.Println("Info: No meta.yaml files found or successfully processed.")
		return
	}

	// Write CSV to stdout
	if err := writeCSV(os.Stdout, entries); err != nil {
		log.Fatalf("Error: Failed to write CSV: %v", err)
	}

	log.Println("Info: CSV generation complete.")
}

// processYAMLFiles finds and processes all meta.yaml files in the directory
func processYAMLFiles(dir string) ([][]string, error) {
	var allRows [][]string
	headers := []string{
		"name", "title", "repo", "synopsis", "year", "start_date", "end_date",
		"attribution_type", "component_type", "lang", "cause", "motive",
		"transitive", "insertion_phase", "impact_type", "impact_user_count",
		"references", "versions", "commits", "historical_artifacts", "current_artifacts",
		"domain", "domain_type", "artifact_type", "hashes",
	}

	// Prepend headers to results
	allRows = append(allRows, headers)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Warning: Error accessing path %q: %v. Skipping.", path, err)
			return nil
		}

		if d.IsDir() || d.Name() != "meta.yaml" {
			return nil
		}

		log.Printf("Info: Processing file: %s", path)
		fileRows, err := yamlFileToRows(path, headers)
		if err != nil {
			log.Printf("Warning: Could not process %s: %v. Skipping.", path, err)
			return nil
		}

		allRows = append(allRows, fileRows...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return allRows[1:], nil // Return data rows (exclude headers)
}

// yamlFileToRows converts a YAML file to CSV rows
func yamlFileToRows(path string, headers []string) ([][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []AttackEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	rows := make([][]string, 0, len(entries))
	for _, entry := range entries {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getFieldValue(entry, header, data)
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// getFieldValue extracts a field value from an AttackEntry
func getFieldValue(entry AttackEntry, field string, originalData []byte) string {
	switch field {
	case "name":
		return entry.Name
	case "title":
		return entry.Title
	case "repo":
		return entry.Repo
	case "synopsis":
		return entry.Synopsis
	case "year":
		return entry.StartDate[0:4]
	case "start_date":
		return entry.StartDate
	case "end_date":
		return entry.EndDate
	case "attribution_type":
		return entry.AttributionType
	case "component_type":
		return entry.ComponentType
	case "lang":
		return entry.Lang
	case "cause":
		return entry.Cause
	case "motive":
		return entry.Motive
	case "transitive":
		return strconv.FormatBool(entry.Transitive)
	case "insertion_phase":
		return entry.InsertionPhase
	case "impact_type":
		return entry.ImpactType
	case "impact_user_count":
		// Check if field is present in original YAML
		if entry.ImpactUserCount == 0 && !strings.Contains(string(originalData), "impact_user_count:") {
			return ""
		}
		return strconv.Itoa(entry.ImpactUserCount)
	case "references":
		return strings.Join(entry.References, ", ")
	case "versions":
		return strings.Join(entry.Versions, ", ")
	case "commits":
		return strings.Join(entry.Commits, ", ")
	case "historical_artifacts":
		return strings.Join(entry.HistoricalArtifacts, ", ")
	case "current_artifacts":
		return strings.Join(entry.CurrentArtifacts, ", ")
	case "domain":
		return entry.Domain
	case "domain_type":
		return entry.DomainType
	case "artifact_type":
		return entry.ArtifactType
	case "hashes":
		return strings.Join(entry.Hashes, ", ")
	default:
		return ""
	}
}

// writeCSV writes data rows to a CSV file
func writeCSV(w *os.File, rows [][]string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write headers
	headers := []string{
		"name", "title", "repo", "synopsis", "start_date", "end_date",
		"attribution_type", "component_type", "lang", "cause", "motive",
		"transitive", "insertion_phase", "impact_type", "impact_user_count",
		"references", "versions", "commits", "historical_artifacts", "current_artifacts",
		"domain", "domain_type", "artifact_type", "hashes",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	// Write data rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			log.Printf("Warning: Error writing row: %v", err)
		}
	}

	return writer.Error()
}
