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

// AttackEntry defines the structure for unmarshalling the meta.yaml files.
// Field names correspond to the schema provided in prompt.txt.
type AttackEntry struct {
	Name            string   `yaml:"name"`
	Title           string   `yaml:"title"`
	Repo            string   `yaml:"repo"`
	Synopsis        string   `yaml:"synopsis"`
	StartDate       string   `yaml:"start_date"`
	EndDate         string   `yaml:"end_date"`
	AttributionType string   `yaml:"attribution_type"`
	ComponentType   string   `yaml:"component_type"`
	Lang            string   `yaml:"lang"`
	Cause           string   `yaml:"cause"`
	Motive          string   `yaml:"motive"`
	Transitive      bool     `yaml:"transitive"`
	InsertionPhase  string   `yaml:"insertion_phase"`
	ImpactType      string   `yaml:"impact_type"`
	ImpactUserCount int      `yaml:"impact_user_count"`
	References      []string `yaml:"references"`
	Versions        []string `yaml:"versions"`
	Commits         []string `yaml:"commits"`
	Artifacts       []string `yaml:"artifacts"`
	Domain          string   `yaml:"domain"`
	DomainType      string   `yaml:"domain_type"`
	ArtifactType    string   `yaml:"artifact_type"`
	ImpactedHashes  []string `yaml:"impacted_hashes"`
}

// csvHeaders defines the exact order and names of columns for the CSV output.
// This order is derived from the prompt.txt specification.
var csvHeaders = []string{
	"name",
	"title",
	"repo",
	"synopsis",
	"start_date",
	"end_date",
	"attribution_type",
	"component_type",
	"lang",
	"cause",
	"motive",
	"transitive",
	"insertion_phase",
	"impact_type",
	"impact_user_count",
	"references",
	"versions",
	"commits",
	"artifacts",
	"domain",
	"domain_type",
	"artifact_type",
	"impacted_hashes",
}

// main is the entry point of the YAML to CSV conversion utility.
// It parses command-line arguments for a search directory,
// finds all 'meta.yaml' files within that directory recursively,
// processes their content, and outputs the aggregated data as a CSV to stdout.
func main() {
	// Determine the root directory for searching 'meta.yaml' files.
	// Defaults to the current directory if no argument is provided.
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalf("Error: Could not determine absolute path for %q: %v", rootDir, err)
	}

	fmt.Fprintf(os.Stderr, "Info: Searching for meta.yaml files in: %s\n", absRootDir)

	// Find and process all 'meta.yaml' files.
	// The headers are now predefined by csvHeaders.
	dataRows, processedHeaders, err := findAndProcessYAMLFiles(absRootDir)
	if err != nil {
		// This error would typically be from WalkDir itself if the initial path is problematic
		// or if the WalkDirFunc decided to propagate a critical error.
		log.Fatalf("Error: Failed to find or process YAML files: %v", err)
	}

	if len(dataRows) == 0 {
		fmt.Fprintln(os.Stderr, "Info: No meta.yaml files found or successfully processed.")
		// Even if no data, we might want to output headers if that's desired for CSVW.
		// For now, exiting if no records.
		// To output headers even with no data, move header writing before this check.
		return
	}

	// Write the collected data to CSV format on standard output.
	if err := writeCSVOutput(os.Stdout, processedHeaders, dataRows); err != nil {
		log.Fatalf("Error: Failed to write CSV output: %v", err)
	}

	fmt.Fprintln(os.Stderr, "Info: CSV generation complete.")
}

// findAndProcessYAMLFiles walks the specified searchDir, identifies 'meta.yaml' files,
// processes each one according to a predefined schema, and aggregates their data.
// It returns a 2D slice of strings (data rows), the predefined CSV headers,
// and an error if the directory walking itself fails critically.
func findAndProcessYAMLFiles(searchDir string) ([][]string, []string, error) {
	var allDataRows [][]string
	// headers are now predefined by the global csvHeaders variable.

	walkErr := filepath.WalkDir(searchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Error accessing path (e.g., permission denied). Log and attempt to skip.
			fmt.Fprintf(os.Stderr, "Warning: Error accessing path %q: %v. Attempting to skip.\n", path, err)
			if d != nil && d.IsDir() {
				return fs.SkipDir // Skip this directory.
			}
			return nil // Skip this file.
		}

		if d.IsDir() || d.Name() != "meta.yaml" {
			return nil // Continue walking, skip directories and non-'meta.yaml' files.
		}

		fmt.Fprintf(os.Stderr, "Info: Processing file: %s\n", path)
		// processMetaFile now returns a [][]string (slice of rows) based on the predefined AttackEntry struct.
		rowsFromFile, procErr := processMetaFile(path)
		if procErr != nil {
			// Error processing a specific file (e.g., malformed YAML, schema mismatch). Log and continue.
			fmt.Fprintf(os.Stderr, "Warning: Could not process file %s: %v. Skipping this file.\\n", path, procErr)
			return nil // Continue walking to process other files.
		}

		if len(rowsFromFile) > 0 {
			allDataRows = append(allDataRows, rowsFromFile...)
		}
		return nil
	})

	if walkErr != nil {
		// This error indicates a failure in WalkDir itself.
		return nil, nil, fmt.Errorf("error walking directory %q: %w", searchDir, walkErr)
	}

	// Headers are now predefined, so no need to build or sort them here.
	// We return the global csvHeaders.
	return allDataRows, csvHeaders, nil
}

// processMetaFile reads a single YAML file specified by filePath,
// unmarshals it into a slice of AttackEntry structs (as a file can contain multiple entries),
// and then maps each struct's fields to a CSV row (slice of strings) based on the predefined csvHeaders.
// It returns a slice of CSV rows ([][]string).
// Errors during file reading or YAML unmarshalling are returned.
func processMetaFile(filePath string) ([][]string, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}

	var entries []AttackEntry // Expect a list of AttackEntry
	if err := yaml.Unmarshal(fileData, &entries); err != nil {
		return nil, fmt.Errorf("unmarshalling YAML from %s: %w", filePath, err)
	}

	var allCsvRows [][]string
	for _, entry := range entries {
		csvRow := make([]string, len(csvHeaders))
		for i, header := range csvHeaders {
			switch header {
			case "name":
				csvRow[i] = entry.Name
			case "title":
				csvRow[i] = entry.Title
			case "repo":
				csvRow[i] = entry.Repo
			case "synopsis":
				csvRow[i] = entry.Synopsis
			case "start_date":
				csvRow[i] = entry.StartDate
			case "end_date":
				csvRow[i] = entry.EndDate
			case "attribution_type":
				csvRow[i] = entry.AttributionType
			case "component_type":
				csvRow[i] = entry.ComponentType
			case "lang":
				csvRow[i] = entry.Lang
			case "cause":
				csvRow[i] = entry.Cause
			case "motive":
				csvRow[i] = entry.Motive
			case "transitive":
				csvRow[i] = strconv.FormatBool(entry.Transitive)
			case "insertion_phase":
				csvRow[i] = entry.InsertionPhase
			case "impact_type":
				csvRow[i] = entry.ImpactType
			case "impact_user_count":
				// For impact_user_count, we check against the original fileData for presence.
				// This check is less precise if multiple entries are in one file and only some omit the field.
				// However, for simplicity, we'll keep it this way. A more complex check would involve
				// re-parsing parts of the YAML or using a more detailed unmarshalling that preserves presence info.
				if entry.ImpactUserCount == 0 && !isFieldPresentInYAML(fileData, "impact_user_count") {
					csvRow[i] = ""
				} else {
					csvRow[i] = strconv.Itoa(entry.ImpactUserCount)
				}
			case "references":
				csvRow[i] = strings.Join(entry.References, ", ")
			case "versions":
				csvRow[i] = strings.Join(entry.Versions, ", ")
			case "commits":
				csvRow[i] = strings.Join(entry.Commits, ", ")
			case "artifacts":
				csvRow[i] = strings.Join(entry.Artifacts, ", ")
			case "domain":
				csvRow[i] = entry.Domain
			case "domain_type":
				csvRow[i] = entry.DomainType
			case "artifact_type":
				csvRow[i] = entry.ArtifactType
			case "impacted_hashes":
				csvRow[i] = strings.Join(entry.ImpactedHashes, ", ")
			default:
				fmt.Fprintf(os.Stderr, "Warning: Unknown header '%s' encountered while processing file %s.\\n", header, filePath)
				csvRow[i] = ""
			}
		}
		allCsvRows = append(allCsvRows, csvRow)
	}
	return allCsvRows, nil
}

// isFieldPresentInYAML checks if a field key is explicitly present in the raw YAML data.
// This is a helper to distinguish between an omitted optional field and a field with a zero value.
// Note: This is a somewhat naive check and might not cover all YAML aliasing/merging complexities,
// but it should work for straightforward key presence.
func isFieldPresentInYAML(yamlData []byte, fieldName string) bool {
	// Check for common YAML key formats: `fieldName:` or `"fieldName":` or `'fieldName':`
	return strings.Contains(string(yamlData), fieldName+":")
}

// writeCSVOutput writes the provided header row and data rows to the given writer (e.g., os.Stdout)
// in CSV format.
// Errors during CSV writing are returned.
func writeCSVOutput(writer *os.File, headerRow []string, dataRows [][]string) error {
	csvWriter := csv.NewWriter(writer)

	// Write the header row.
	if err := csvWriter.Write(headerRow); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	// Write data rows.
	for i, row := range dataRows {
		if err := csvWriter.Write(row); err != nil {
			// Log error for a specific row but attempt to continue.
			// The final csvWriter.Error() will catch persistent issues.
			fmt.Fprintf(os.Stderr, "Warning: Error writing CSV row %d for data %v: %v\n", i+1, row, err)
		}
	}

	// Flush any buffered data to the writer.
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		// This checks for errors that occurred during any Write or during Flush.
		return fmt.Errorf("flushing CSV writer: %w", err)
	}
	return nil
}
