package main

import (
	"encoding/csv"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	// Configure logging to stderr
	log.SetPrefix("")
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

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

	// Process YAML files and generate CSV data
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write headers to CSV
	if err := writer.Write(csvHeaders); err != nil {
		log.Fatalf("Error: Failed to write CSV headers: %v", err)
	}

	// Walk directory tree searching for meta.yaml files
	processCount := 0
	err = filepath.WalkDir(absRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Warning: Error accessing path %q: %v. Skipping.", path, err)
			return nil
		}

		if d.IsDir() || d.Name() != "meta.yaml" {
			return nil
		}

		log.Printf("Info: Processing file: %s", path)
		
		// Read and parse YAML file
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: Could not read %s: %v. Skipping.", path, err)
			return nil
		}

		var entries []AttackEntry
		if err := yaml.Unmarshal(data, &entries); err != nil {
			log.Printf("Warning: Could not process %s: %v. Skipping.", path, err)
			return nil
		}

		// Convert entries to CSV rows and write directly
		for _, entry := range entries {
			row := make([]string, len(csvHeaders))
			
			// Extract values for each field
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
					} else {
						row[i] = ""
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
					// Only output non-zero values or explicit zeros
					if entry.ImpactUserCount != 0 || strings.Contains(string(data), "impact_user_count:") {
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
			
			// Write row directly to CSV
			if err := writer.Write(row); err != nil {
				log.Printf("Warning: Error writing row: %v", err)
				continue
			}
			
			processCount++
		}
		
		return nil
	})

	if err != nil {
		log.Fatalf("Error: Failed walking directory: %v", err)
	}

	if processCount == 0 {
		log.Println("Info: No meta.yaml files found or successfully processed.")
		return
	}

	log.Println("Info: CSV generation complete.")
}