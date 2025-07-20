package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestCase represents a single test file
type TestCase struct {
	Name  string // "precedence.lox"
	Path  string // full file path
	Suite string // "print", "class", or "" for top-level
}

// ExecutionResult holds results from executing a lox program
type ExecutionResult struct {
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

// TestResult holds comparison results between reference and target
type TestResult struct {
	TestCase  *TestCase
	Reference ExecutionResult
	Target    ExecutionResult
	Passed    bool
	Errors    []string
}

// TestFramework is the main testing framework
type TestFramework struct {
	CasesDir     string
	ReferenceCmd string
	TargetCmd    string
	TestCases    []*TestCase
	Results      []*TestResult
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  discover           - Discover all test cases")
		fmt.Println("  collect            - Collect reference results for all tests")
		fmt.Println("  test <target_cmd>  - Run tests against target implementation")
		fmt.Println("  benchmark <target_cmd> - Compare execution speeds")
		return
	}

	framework := NewTestFramework("cases", "./official-clox", "")
	command := os.Args[1]

	switch command {
	case "discover":
		if err := framework.DiscoverTests(); err != nil {
			fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
			os.Exit(1)
		}
		framework.PrintDiscovery()

	case "collect":
		if err := framework.DiscoverTests(); err != nil {
			fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
			os.Exit(1)
		}
		if err := framework.CollectReference(); err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting reference results: %v\n", err)
			os.Exit(1)
		}

	case "test":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: go run main.go test <target_command>\n")
			os.Exit(1)
		}
		framework.TargetCmd = os.Args[2]

		if err := framework.DiscoverTests(); err != nil {
			fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
			os.Exit(1)
		}
		if err := framework.RunTests(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running tests: %v\n", err)
			os.Exit(1)
		}
		framework.PrintResults()

	case "benchmark":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: go run main.go benchmark <target_command>\n")
			os.Exit(1)
		}
		framework.TargetCmd = os.Args[2]

		if err := framework.DiscoverTests(); err != nil {
			fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
			os.Exit(1)
		}
		if err := framework.RunBenchmarks(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running benchmarks: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func NewTestFramework(casesDir, referenceCmd, targetCmd string) *TestFramework {
	return &TestFramework{
		CasesDir:     casesDir,
		ReferenceCmd: referenceCmd,
		TargetCmd:    targetCmd,
	}
}

func (tf *TestFramework) DiscoverTests() error {
	tf.TestCases = nil

	return filepath.WalkDir(tf.CasesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".lox") {
			return nil
		}

		relPath, _ := filepath.Rel(tf.CasesDir, path)
		suite := filepath.Dir(relPath)
		if suite == "." {
			suite = ""
		}

		testCase := &TestCase{
			Name:  d.Name(),
			Path:  path,
			Suite: suite,
		}

		tf.TestCases = append(tf.TestCases, testCase)
		return nil
	})
}

func (tf *TestFramework) CollectReference() error {
	fmt.Printf("Collecting reference results using: %s\n", tf.ReferenceCmd)
	fmt.Printf("Processing %d tests...\n", len(tf.TestCases))

	// Ensure reference results directory exists
	if err := os.MkdirAll("reference_results", 0755); err != nil {
		return err
	}

	for i, test := range tf.TestCases {
		fmt.Printf("  [%d/%d] %s", i+1, len(tf.TestCases), test.Name)
		if test.Suite != "" {
			fmt.Printf(" (%s)", test.Suite)
		}
		fmt.Print("...")

		result, err := tf.executeTest(tf.ReferenceCmd, test)
		if err != nil {
			fmt.Printf(" ERROR: %v\n", err)
			continue
		}

		if err := tf.saveReferenceResult(test, result); err != nil {
			fmt.Printf(" ERROR saving: %v\n", err)
			continue
		}

		fmt.Printf(" OK (%.2fms)\n", float64(result.Duration.Nanoseconds())/1000000)
	}

	return nil
}

func (tf *TestFramework) RunTests() error {
	fmt.Printf("Running %d tests against: %s\n", len(tf.TestCases), tf.TargetCmd)

	tf.Results = make([]*TestResult, 0, len(tf.TestCases))

	for i, test := range tf.TestCases {
		fmt.Printf("  [%d/%d] %s", i+1, len(tf.TestCases), test.Name)
		if test.Suite != "" {
			fmt.Printf(" (%s)", test.Suite)
		}
		fmt.Print("...")

		result := tf.compareTest(test)
		tf.Results = append(tf.Results, result)

		if result.Passed {
			fmt.Printf(" PASS")
		} else {
			fmt.Printf(" FAIL")
		}
		fmt.Printf(" (%.2fms)\n", float64(result.Target.Duration.Nanoseconds())/1000000)
	}

	return nil
}

func (tf *TestFramework) compareTest(test *TestCase) *TestResult {
	result := &TestResult{
		TestCase: test,
		Passed:   false,
	}

	// Load reference result
	refResult, err := tf.loadReferenceResult(test)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load reference: %v", err))
		return result
	}
	result.Reference = *refResult

	// Execute target
	targetResult, err := tf.executeTest(tf.TargetCmd, test)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to execute target: %v", err))
		return result
	}
	result.Target = *targetResult

	// Compare results
	result.Passed = true

	if result.Reference.ExitCode != result.Target.ExitCode {
		result.Passed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("Exit code mismatch: expected %d, got %d",
				result.Reference.ExitCode, result.Target.ExitCode))
	}

	if strings.TrimSpace(result.Reference.Stdout) != strings.TrimSpace(result.Target.Stdout) {
		result.Passed = false
		result.Errors = append(result.Errors, "Stdout mismatch")
	}

	if strings.TrimSpace(result.Reference.Stderr) != strings.TrimSpace(result.Target.Stderr) {
		result.Passed = false
		result.Errors = append(result.Errors, "Stderr mismatch")
	}

	return result
}

func (tf *TestFramework) executeTest(cmd string, test *TestCase) (*ExecutionResult, error) {
	start := time.Now()

	// Parse command
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// Execute: cmd test.Path
	cmdArgs := append(parts[1:], test.Path)
	execCmd := exec.Command(parts[0], cmdArgs...)

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("execution error: %w", err)
		}
	}

	return &ExecutionResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

func (tf *TestFramework) saveReferenceResult(test *TestCase, result *ExecutionResult) error {
	filename := tf.getReferenceFileName(test)
	path := filepath.Join("reference_results", filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (tf *TestFramework) loadReferenceResult(test *TestCase) (*ExecutionResult, error) {
	filename := tf.getReferenceFileName(test)
	path := filepath.Join("reference_results", filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reference result not found (run 'collect' first): %w", err)
	}

	var result ExecutionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (tf *TestFramework) getReferenceFileName(test *TestCase) string {
	filename := strings.ReplaceAll(test.Name, ".lox", ".json")
	if test.Suite != "" {
		filename = strings.ReplaceAll(test.Suite, "/", "_") + "_" + filename
	}
	return filename
}

func (tf *TestFramework) RunBenchmarks() error {
	fmt.Printf("Benchmarking %d tests\n", len(tf.TestCases))
	fmt.Printf("Reference: %s\n", tf.ReferenceCmd)
	fmt.Printf("Target: %s\n", tf.TargetCmd)
	fmt.Println()

	totalRefTime := time.Duration(0)
	totalTargetTime := time.Duration(0)

	fmt.Printf("%-40s %12s %12s %8s\n", "Test", "Reference", "Target", "Ratio")
	fmt.Println(strings.Repeat("-", 80))

	for _, test := range tf.TestCases {
		testName := test.Name
		if test.Suite != "" {
			testName = test.Suite + "/" + test.Name
		}

		// Run reference
		refResult, err := tf.executeTest(tf.ReferenceCmd, test)
		if err != nil {
			fmt.Printf("%-40s %12s %12s %8s\n", testName, "ERROR", "", "")
			continue
		}

		// Run target
		targetResult, err := tf.executeTest(tf.TargetCmd, test)
		if err != nil {
			fmt.Printf("%-40s %12s %12s %8s\n", testName,
				fmt.Sprintf("%.2fms", float64(refResult.Duration.Nanoseconds())/1000000),
				"ERROR", "")
			continue
		}

		totalRefTime += refResult.Duration
		totalTargetTime += targetResult.Duration

		refMs := float64(refResult.Duration.Nanoseconds()) / 1000000
		targetMs := float64(targetResult.Duration.Nanoseconds()) / 1000000
		ratio := targetMs / refMs

		fmt.Printf("%-40s %10.2fms %10.2fms %7.2fx\n",
			testName, refMs, targetMs, ratio)
	}

	fmt.Println(strings.Repeat("-", 80))
	totalRefMs := float64(totalRefTime.Nanoseconds()) / 1000000
	totalTargetMs := float64(totalTargetTime.Nanoseconds()) / 1000000
	totalRatio := totalTargetMs / totalRefMs

	fmt.Printf("%-40s %10.2fms %10.2fms %7.2fx\n",
		"TOTAL", totalRefMs, totalTargetMs, totalRatio)

	return nil
}

func (tf *TestFramework) PrintDiscovery() {
	suiteCount := make(map[string]int)
	topLevelCount := 0

	for _, test := range tf.TestCases {
		if test.Suite == "" {
			topLevelCount++
		} else {
			suiteCount[test.Suite]++
		}
	}

	fmt.Printf("Discovered %d test cases:\n", len(tf.TestCases))

	if topLevelCount > 0 {
		fmt.Printf("  Top-level: %d tests\n", topLevelCount)
	}

	for suite, count := range suiteCount {
		fmt.Printf("  %s: %d tests\n", suite, count)
	}
}

func (tf *TestFramework) PrintResults() {
	passed := 0
	failed := 0

	for _, result := range tf.Results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("\nTest Results: %d passed, %d failed\n", passed, failed)

	if failed > 0 {
		fmt.Println("\nFailures:")
		for _, result := range tf.Results {
			if !result.Passed {
				testName := result.TestCase.Name
				if result.TestCase.Suite != "" {
					testName = result.TestCase.Suite + "/" + testName
				}

				fmt.Printf("\n%s:\n", testName)
				for _, err := range result.Errors {
					fmt.Printf("  %s\n", err)
				}

				// Show detailed comparison for mismatches
				if len(result.Errors) == 0 || containsMatchError(result.Errors) {
					if result.Reference.Stdout != result.Target.Stdout {
						fmt.Printf("  Expected stdout: %q\n", result.Reference.Stdout)
						fmt.Printf("  Actual stdout:   %q\n", result.Target.Stdout)
					}
					if result.Reference.Stderr != result.Target.Stderr {
						fmt.Printf("  Expected stderr: %q\n", result.Reference.Stderr)
						fmt.Printf("  Actual stderr:   %q\n", result.Target.Stderr)
					}
					if result.Reference.ExitCode != result.Target.ExitCode {
						fmt.Printf("  Expected exit:   %d\n", result.Reference.ExitCode)
						fmt.Printf("  Actual exit:     %d\n", result.Target.ExitCode)
					}
				}
			}
		}
	}
}

func containsMatchError(errors []string) bool {
	for _, err := range errors {
		if strings.Contains(err, "mismatch") {
			return true
		}
	}
	return false
}
