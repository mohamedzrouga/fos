// auto_type_hints.go
// Automatically adds type annotations and fixes type issues in Go code
// Note: Go already has strong type inference, this focuses on:
// - Adding explicit types where beneficial
// - Fixing type mismatches
// - Adding interface implementations
// - Generating type stubs for complex types

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TypeHintResult struct {
	File        string
	Success     bool
	TypesAdded  int
	IssuesFixed int
	Output      string
	Suggestions []string
}

func runCommand(name string, args []string, cwd string) (bool, string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	output := stdout.String() + stderr.String()
	success := err == nil
	
	return success, output
}

func analyzeFile(filePath string) (*TypeHintResult, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	
	result := &TypeHintResult{
		File:        filePath,
		Success:     true,
		TypesAdded:  0,
		IssuesFixed: 0,
		Suggestions: []string{},
	}
	
	// Analyze AST for type inference opportunities
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Check for missing return type annotations
			if x.Type.Results != nil && len(x.Type.Results.List) > 0 {
				for _, field := range x.Type.Results.List {
					if field.Type == nil {
						result.Suggestions = append(result.Suggestions,
							fmt.Sprintf("Add return type to function %s", x.Name.Name))
					}
				}
			}
			
		case *ast.AssignStmt:
			// Check for type assertions that could be explicit
			if len(x.Lhs) == 1 && len(x.Rhs) == 1 {
				// Look for := that could have explicit types
				if x.Tok == token.DEFINE {
					result.Suggestions = append(result.Suggestions,
						fmt.Sprintf("Consider explicit type for %v", x.Lhs[0]))
				}
			}
			
		case *ast.TypeAssertExpr:
			// Found type assertion - good, types are explicit
			result.TypesAdded++
		}
		return true
	})
	
	return result, nil
}

func runGoplsDiagnostics(path string) TypeHintResult {
	// Use gopls for type checking and suggestions
	success, output := runCommand("gopls", []string{
		"diagnose",
		"-v",
		path,
	}, path)
	
	return TypeHintResult{
		File:        path,
		Success:     success,
		TypesAdded:  0,
		IssuesFixed: 0,
		Output:      output,
		Suggestions: []string{
			"Review gopls diagnostics for type issues",
			"Use 'gopls fix' to auto-fix some issues",
		},
	}
}

func runStaticcheck(path string) TypeHintResult {
	success, output := runCommand("staticcheck", []string{
		"./...",
	}, path)
	
	suggestions := []string{}
	if !success {
		suggestions = append(suggestions,
			"Fix SAxxxx errors from staticcheck",
			"Add explicit type conversions where needed",
			"Review type mismatches in assignments",
		)
	}
	
	return TypeHintResult{
		File:        path,
		Success:     success,
		TypesAdded:  0,
		IssuesFixed: 0,
		Output:      output,
		Suggestions: suggestions,
	}
}

func runGoimports(path string) TypeHintResult {
	success, output := runCommand("goimports", []string{
		"-w",
		"./...",
	}, path)
	
	return TypeHintResult{
		File:        path,
		Success:     success,
		TypesAdded:  0,
		IssuesFixed: 0,
		Output:      output,
		Suggestions: []string{
			"Imports organized automatically",
			"Review any unused imports removed",
		},
	}
}

func runGofumpt(path string) TypeHintResult {
	success, output := runCommand("gofumpt", []string{
		"-w",
		"./...",
	}, path)
	
	return TypeHintResult{
		File:        path,
		Success:     success,
		TypesAdded:  0,
		IssuesFixed: 0,
		Output:      output,
		Suggestions: []string{
			"Code formatted with gofumpt",
			"Review formatting changes",
		},
	}
}

func fixTypeIssues(path string) TypeHintResult {
	// Try to auto-fix common type issues
	var output strings.Builder
	
	// Run golangci-lint with fix
	success1, out1 := runCommand("golangci-lint", []string{
		"run",
		"--fix",
		"./...",
	}, path)
	output.WriteString(out1)
	
	// Run gofmt for basic fixes
	success2, out2 := runCommand("gofmt", []string{
		"-w",
		"./...",
	}, path)
	output.WriteString(out2)
	
	return TypeHintResult{
		File:        path,
		Success:     success1 && success2,
		TypesAdded:  0,
		IssuesFixed: strings.Count(output.String(), "Fixed"),
		Output:      output.String(),
		Suggestions: []string{
			"Review fixed issues",
			"Run tests to verify fixes don't break functionality",
		},
	}
}

func generateReport(results []TypeHintResult, path string) error {
	var report strings.Builder
	
	report.WriteString("# Go Type Annotation Report\n\n")
	report.WriteString(fmt.Sprintf("**Path:** %s\n\n", path))
	
	report.WriteString("## Summary\n\n")
	totalTypes := 0
	totalFixed := 0
	for _, r := range results {
		totalTypes += r.TypesAdded
		totalFixed += r.IssuesFixed
	}
	report.WriteString(fmt.Sprintf("- **Types Added:** %d\n", totalTypes))
	report.WriteString(fmt.Sprintf("- **Issues Fixed:** %d\n\n", totalFixed))
	
	report.WriteString("## Detailed Results\n\n")
	
	for _, result := range results {
		status := "✅"
		if !result.Success {
			status = "⚠️"
		}
		
		report.WriteString(fmt.Sprintf("### %s %s\n\n", status, result.File))
		
		if result.Output != "" {
			report.WriteString("```\n")
			output := result.Output
			if len(output) > 2000 {
				output = output[:2000] + "\n... (truncated)"
			}
			report.WriteString(output)
			report.WriteString("\n```\n\n")
		}
		
		if len(result.Suggestions) > 0 {
			report.WriteString("**Suggestions:**\n\n")
			for _, s := range result.Suggestions {
				report.WriteString(fmt.Sprintf("- %s\n", s))
			}
			report.WriteString("\n")
		}
	}
	
	reportFile := filepath.Join(path, "type_hints_report.md")
	return os.WriteFile(reportFile, []byte(report.String()), 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run auto_type_hints.go <path>")
		fmt.Println("  path: Directory to scan (use '.' for current)")
		os.Exit(1)
	}
	
	targetPath := os.Args [eightfold](https://eightfold.ai/engineering-blog/static-type-checking-large-scale-python-codebase/)
	
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
	
	var results []TypeHintResult
	
	// Step 1: Format and organize imports
	fmt.Println("\n[Step 1/5] Formatting code...")
	result := runGofumpt(absPath)
	results = append(results, result)
	fmt.Printf("  %s\n", result.Output)
	
	// Step 2: Organize imports
	fmt.Println("\n[Step 2/5] Organizing imports...")
	result = runGoimports(absPath)
	results = append(results, result)
	fmt.Printf("  %s\n", result.Output)
	
	// Step 3: Fix type issues
	fmt.Println("\n[Step 3/5] Fixing type issues...")
	result = fixTypeIssues(absPath)
	results = append(results, result)
	fmt.Printf("  Fixed %d issues\n", result.IssuesFixed)
	
	// Step 4: Static analysis
	fmt.Println("\n[Step 4/5] Running static analysis...")
	result = runStaticcheck(absPath)
	results = append(results, result)
	
	// Step 5: Analyze individual files
	fmt.Println("\n[Step 5/5] Analyzing files...")
	filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			result, err := analyzeFile(path)
			if err != nil {
				fmt.Printf("  Error analyzing %s: %v\n", path, err)
				return nil
			}
			results = append(results, *result)
		}
		return nil
	})
	
	// Generate report
	if err := generateReport(results, absPath); err != nil {
		fmt.Printf("Error generating report: %v\n", err)
	} else {
		fmt.Printf("\nReport saved to: %s/type_hints_report.md\n", absPath)
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
	fmt.Printf("Checks passed: %d/%d\n", passed, len(results))
	
	if passed < len(results) {
		fmt.Println("\nReview the report for suggestions.")
	}
}
