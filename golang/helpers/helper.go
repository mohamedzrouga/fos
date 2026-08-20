// scan_code.go
// Code Quality Scanner for Go
// Scans for: linting, type checking, performance, security, best practices

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ScanResult struct {
	Tool       string   `json:"tool"`
	Success    bool     `json:"success"`
	Output     string   `json:"output"`
	Suggestions []string `json:"suggestions"`
	Duration   string   `json:"duration"`
}

type Report struct {
	Generated string       `json:"generated"`
	Path      string       `json:"path"`
	Results   []ScanResult `json:"results"`
	Summary   Summary      `json:"summary"`
}

type Summary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

func runCommand(name string, args []string, cwd string) (bool, string, time.Duration) {
	start := time.Now()
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	duration := time.Since(start)
	
	output := stdout.String() + stderr.String()
	success := err == nil
	
	return success, output, duration
}

func checkGolangciLint(path string) ScanResult {
	start := time.Now()
	
	// Run golangci-lint with comprehensive linters
	success, output, _ := runCommand("golangci-lint", []string{
		"run",
		"--enable-all",
		"--disable", "gomnd", // Too noisy
		"--disable", "testpackage", // Optional
		"--max-same-issues", "50",
		"--timeout", "5m",
		"--out-format", "colored-line-number",
		path,
	}, path)
	
	suggestions := []string{}
	if !success {
		suggestions = append(suggestions,
			"Run 'golangci-lint run --fix' to auto-fix issues",
			"Configure linters in .golangci.yml",
			"Use '//nolint:linter' for intentional violations",
		)
	}
	
	return ScanResult{
		Tool:        "golangci-lint (comprehensive)",
		Success:     success,
		Output:      output,
		Suggestions: suggestions,
		Duration:    time.Since(start).String(),
	}
}

func checkGoVet(path string) ScanResult {
	start := time.Now()
	success, output, _ := runCommand("go", []string{"vet", "./..."}, path)
	
	suggestions := []string{}
	if !success {
		suggestions = append(suggestions,
			"Check for unreachable code",
			"Verify printf format strings",
			"Review suspicious constructs",
		)
	}
	
	return ScanResult{
		Tool:        "go vet",
		Success:     success,
		Output:      output,
		Suggestions: suggestions,
		Duration:    time.Since(start).String(),
	}
}

func checkStaticCheck(path string) ScanResult {
	start := time.Now()
	success, output, _ := runCommand("staticcheck", []string{"./..."}, path)
	
	suggestions := []string{}
	if !success {
		suggestions = append(suggestions,
			"Review SAxxxx codes at https://staticcheck.dev/docs/checks",
			"Fix nil pointer dereferences",
			"Remove unused code",
		)
	}
	
	return ScanResult{
		Tool:        "staticcheck",
		Success:     success,
		Output:      output,
		Suggestions: suggestions,
		Duration:    time.Since(start).String(),
	}
}

func checkGoSec(path string) ScanResult {
	start := time.Now()
	success, output, _ := runCommand("gosec", []string{
		"-exclude-dir", "vendor",
		"-exclude-dir", "testdata",
		"./...",
	}, path)
	
	suggestions := []string{}
	if !success {
		suggestions = append(suggestions,
			"Avoid hardcoded credentials",
			"Use crypto/rand instead of math/rand for security",
			"Validate TLS configurations",
			"Sanitize SQL queries",
		)
	}
	
	return ScanResult{
		Tool:        "gosec (security)",
		Success:     success,
		Output:      output,
		Suggestions: suggestions,
		Duration:    time.Since(start).String(),
	}
}

func checkPerf(path string) ScanResult {
	start := time.Now()
	
	// Run multiple performance-focused checks
	var output strings.Builder
	
	// ineffassign: unused assignments
	success1, out1, _ := runCommand("ineffassign", []string{"./..."}, path)
	output.WriteString("=== ineffassign ===\n")
	output.WriteString(out1)
	
	// unparam: unused parameters
	success2, out2, _ := runCommand("unparam", []string{"./..."}, path)
	output.WriteString("\n=== unparam ===\n")
	output.WriteString(out2)
	
	success := success1 && success2
	
	suggestions := []string{}
	if !success {
		suggestions = append(suggestions,
			"Remove unused variable assignments",
			"Remove unused function parameters",
			"Consider using 'var _ = x' for intentional unused values",
		)
	}
	
	return ScanResult{
		Tool:        "Performance checks (ineffassign, unparam)",
		Success:     success,
		Output:      output.String(),
		Suggestions: suggestions,
		Duration:    time.Since(start).String(),
	}
}

func checkCyclo(path string) ScanResult {
	start := time.Now()
	success, output, _ := runCommand("gocyclo", []string{
		"-over", "15",
		"./...",
	}, path)
	
	suggestions := []string{}
	if !success {
		suggestions = append(suggestions,
			"Refactor functions with high cyclomatic complexity (>15)",
			"Break down large functions into smaller units",
			"Use early returns to reduce nesting",
		)
	}
	
	return ScanResult{
		Tool:        "gocyclo (complexity)",
		Success:     success,
		Output:      output,
		Suggestions: suggestions,
		Duration:    time.Since(start).String(),
	}
}

func generateReport(results []ScanResult, path string) error {
	report := Report{
		Generated: time.Now().Format(time.RFC3339),
		Path:      path,
		Results:   results,
	}
	
	passed := 0
	for _, r := range results {
		if r.Success {
			passed++
		}
	}
	report.Summary = Summary{
		Total:  len(results),
		Passed: passed,
		Failed: len(results) - passed,
	}
	
	// JSON report
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	jsonFile := filepath.Join(path, "code_quality_report.json")
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return err
	}
	
	// Markdown report
	var md strings.Builder
	md.WriteString("# Code Quality Scan Report\n\n")
	md.WriteString(fmt.Sprintf("**Generated:** %s\n", report.Generated))
	md.WriteString(fmt.Sprintf("**Path:** %s\n\n", report.Path))
	
	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Checks:** %d\n", report.Summary.Total))
	md.WriteString(fmt.Sprintf("- **Passed:** %d\n", report.Summary.Passed))
	md.WriteString(fmt.Sprintf("- **Failed:** %d\n\n", report.Summary.Failed))
	
	md.WriteString("## Detailed Results\n\n")
	
	for _, result := range results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ ISSUES FOUND"
		}
		
		md.WriteString(fmt.Sprintf("### %s: %s (%s)\n\n", result.Tool, status, result.Duration))
		
		if result.Output != "" {
			md.WriteString("```\n")
			// Truncate long outputs
			output := result.Output
			if len(output) > 3000 {
				output = output[:3000] + "\n... (truncated)"
			}
			md.WriteString(output)
			md.WriteString("\n```\n\n")
		}
		
		if len(result.Suggestions) > 0 {
			md.WriteString("**Suggestions:**\n\n")
			for _, suggestion := range result.Suggestions {
				md.WriteString(fmt.Sprintf("- %s\n", suggestion))
			}
			md.WriteString("\n")
		}
	}
	
	mdFile := filepath.Join(path, "code_quality_report.md")
	return os.WriteFile(mdFile, []byte(md.String()), 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scan_code.go <path>")
		fmt.Println("  path: Directory to scan (use '.' for current)")
		os.Exit(1)
	}
	
	targetPath := os.Args[1]
	
	// Resolve to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}
	
	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Printf("Error: Path does not exist: %s\n", absPath)
		os.Exit(1)
	}
	
	if !info.IsDir() {
		fmt.Printf("Error: Path is not a directory: %s\n", absPath)
		os.Exit(1)
	}
	
	fmt.Printf("Scanning: %s\n", absPath)
	fmt.Println(strings.Repeat("=", 60))
	
	var results []ScanResult
	
	checks := []struct {
		name string
		fn   func(string) ScanResult
	}{
		{"golangci-lint", checkGolangciLint},
		{"go vet", checkGoVet},
		{"staticcheck", checkStaticCheck},
		{"gosec (security)", checkGoSec},
		{"Performance", checkPerf},
		{"Complexity", checkCyclo},
	}
	
	for _, check := range checks {
		fmt.Printf("\nRunning %s...\n", check.name)
		result := check.fn(absPath)
		results = append(results, result)
		
		status := "✅ PASS"
		if !result.Success {
			status = "⚠️  ISSUES"
		}
		fmt.Printf("  %s (%s)\n", status, result.Duration)
	}
	
	// Generate report
	reportFile := filepath.Join(absPath, "code_quality_report.md")
	if err := generateReport(results, absPath); err != nil {
		fmt.Printf("Error generating report: %v\n", err)
	} else {
		fmt.Printf("\nReport saved to: %s\n", reportFile)
	}
	
	// Summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SUMMARY")
	passed := 0
	for _, r := range results {
		if r.Success {
			passed++
		}
	}
	fmt.Printf("Passed: %d/%d\n", passed, len(results))
	
	if passed < len(results) {
		fmt.Println("\nReview the report for detailed suggestions.")
		os.Exit(1)
	}
}
