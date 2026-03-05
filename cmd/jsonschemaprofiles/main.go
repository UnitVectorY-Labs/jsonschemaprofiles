package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	jsp "github.com/UnitVectorY-Labs/jsonschemaprofiles"
)

var Version = "dev"

// Exit codes
const (
	exitSuccess  = 0
	exitInvalid  = 1
	exitInternal = 2
)

func main() {
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitInternal)
	}

	switch os.Args[1] {
	case "profiles":
		if len(os.Args) >= 3 && os.Args[2] == "list" {
			cmdProfilesList()
		} else {
			fmt.Fprintf(os.Stderr, "Usage: jsonschemaprofiles profiles list\n")
			os.Exit(exitInternal)
		}
	case "validate":
		if len(os.Args) >= 3 && os.Args[2] == "schema" {
			cmdValidateSchema(os.Args[3:])
		} else {
			fmt.Fprintf(os.Stderr, "Usage: jsonschemaprofiles validate schema [flags]\n")
			os.Exit(exitInternal)
		}
	case "coerce":
		if len(os.Args) >= 3 && os.Args[2] == "schema" {
			cmdCoerceSchema(os.Args[3:])
		} else {
			fmt.Fprintf(os.Stderr, "Usage: jsonschemaprofiles coerce schema [flags]\n")
			os.Exit(exitInternal)
		}
	case "version":
		fmt.Printf("jsonschemaprofiles %s\n", Version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(exitInternal)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `jsonschemaprofiles - Validate JSON Schemas against provider profiles

Usage:
  jsonschemaprofiles profiles list              List available profiles
  jsonschemaprofiles validate schema [flags]    Validate a schema against a profile
  jsonschemaprofiles coerce schema [flags]      Coerce a schema for profile compliance
  jsonschemaprofiles version                    Show version

`)
}

func cmdProfilesList() {
	profiles := jsp.ListProfiles()
	for _, p := range profiles {
		fmt.Printf("%-25s %s\n", p.ID, p.Title)
		fmt.Printf("%-25s %s\n", "", p.Description)
		fmt.Printf("%-25s Baseline: %s  Schema: %s\n\n", "", p.Baseline, p.SchemaFile)
	}
}

func cmdValidateSchema(args []string) {
	fs := flag.NewFlagSet("validate schema", flag.ExitOnError)
	profile := fs.String("profile", "", "Profile ID (required)")
	inputFile := fs.String("in", "", "Input schema file (or - for stdin)")
	format := fs.String("format", "text", "Output format: text or json")
	strict := fs.Bool("strict", false, "Treat warnings as errors")
	modelTarget := fs.String("model-target", "", "Model target (e.g., fine-tuned)")

	if err := fs.Parse(args); err != nil {
		os.Exit(exitInternal)
	}

	if *profile == "" {
		fmt.Fprintf(os.Stderr, "Error: --profile is required\n")
		os.Exit(exitInternal)
	}

	schemaBytes, err := readInput(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(exitInternal)
	}

	opts := &jsp.ValidateOptions{
		Strict:      *strict,
		ModelTarget: *modelTarget,
	}

	report, err := jsp.ValidateSchema(jsp.ProfileID(*profile), schemaBytes, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitInternal)
	}

	outputReport(report, *format)

	if !report.Valid {
		os.Exit(exitInvalid)
	}
}

func cmdCoerceSchema(args []string) {
	fs := flag.NewFlagSet("coerce schema", flag.ExitOnError)
	profile := fs.String("profile", "", "Profile ID (required)")
	inputFile := fs.String("in", "", "Input schema file (or - for stdin)")
	outputFile := fs.String("out", "", "Output schema file (default: stdout)")
	dryRun := fs.Bool("dry-run", false, "Show proposed changes without applying")
	mode := fs.String("mode", "conservative", "Coercion mode: conservative or permissive")
	format := fs.String("format", "text", "Report format: text or json")

	if err := fs.Parse(args); err != nil {
		os.Exit(exitInternal)
	}

	if *profile == "" {
		fmt.Fprintf(os.Stderr, "Error: --profile is required\n")
		os.Exit(exitInternal)
	}

	schemaBytes, err := readInput(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(exitInternal)
	}

	opts := &jsp.CoerceOptions{
		Mode:   jsp.CoerceMode(*mode),
		DryRun: *dryRun,
	}

	coercedBytes, report, changed, err := jsp.CoerceSchema(jsp.ProfileID(*profile), schemaBytes, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitInternal)
	}

	// Write coerced schema
	if !*dryRun && changed {
		if *outputFile != "" {
			if err := os.WriteFile(*outputFile, coercedBytes, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
				os.Exit(exitInternal)
			}
		} else {
			os.Stdout.Write(coercedBytes)
			fmt.Println()
		}
	}

	// Print report to stderr
	outputReportTo(report, *format, os.Stderr)

	if !report.Valid {
		os.Exit(exitInvalid)
	}
}

func readInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func outputReport(report *jsp.Report, format string) {
	outputReportTo(report, format, os.Stdout)
}

func outputReportTo(report *jsp.Report, format string, w io.Writer) {
	switch strings.ToLower(format) {
	case "json":
		b, err := report.JSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error serializing report: %v\n", err)
			return
		}
		fmt.Fprintln(w, string(b))
	default:
		fmt.Fprint(w, report.Text())
	}
}
