package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"
)

type TestCase struct {
	Name     string
	Expected *TestResult
	Actual   *TestResult
	Percent  float64
}

type TestResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

type TestSuite struct {
	Name  string
	Cases []TestCase
}

type TestFramework struct {
	Reference string //command to run the reference implementation
	Target    string //command to run the implementation being tested
	Suites    []*TestSuite
	Total     int
	Failed    []*TestCase
	Percent   float64 //percent difference time to run
}

var (
	noFailStderr = flag.Bool("no-fail-stderr", false, "Stderr mis-match is not a failure.")
)

func main() {
	flag.Parse()

	tf := TestFramework{
		Reference: "test/official-clox",
		Target:    "clox/clox_interpreter",
	}

	tf.collectSuites("test/cases")
	slices.SortFunc(tf.Suites, func(a, b *TestSuite) int {
		return strings.Compare(a.Name, b.Name)
	})

	tf.executeTests()
	tf.PrintSummary()
}

/* Collect the tests from the files and directories in test/cases
 * These only collect one level deeper for a test suite; there are no nested
 * test suites.
 */
func (tf *TestFramework) collectSuites(dir string) {
	suites := []*TestSuite{}
	topLevel := TestSuite{Name: "Top Level"}

	for _, entry := range getEntries(dir) {
		if entry.IsDir() {
			suitePath := path.Join(dir, entry.Name())
			suites = append(suites, collectSuite(suitePath))
		} else {
			topLevel.Cases = append(topLevel.Cases, TestCase{Name: entry.Name()})
		}
	}

	suites = append(suites, &topLevel)
	tf.Suites = suites
}

func getEntries(dir string) []fs.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		os.Exit(1)
	}
	return entries
}

func collectSuite(dir string) *TestSuite {
	suite := &TestSuite{Name: path.Base(dir)}
	for _, entry := range getEntries(dir) {
		if !entry.IsDir() {
			suite.Cases = append(suite.Cases, TestCase{Name: entry.Name()})
		}
	}
	return suite
}

/* These run the tests. It ignores the test in the benchmark test suite because
 * those tests print out how long the test took, which even using the same VM
 * will produce different results.
 */
const WIDTH = 120

func (tf *TestFramework) executeTests() {
	first := true

	for _, suite := range tf.Suites {
		if suite.Name == "benchmark" {
			continue
			// The benchmarks print how long they take, so they will always fail to have
			// the same output
		}

		if first {
			first = false
		} else {
			fmt.Println()
		}

		// Width of 9 for percent to take into account the '%'
		columns := fmt.Sprintf("%12s %12s %8s", "reference", "actual", "percent")
		spacing := strings.Repeat(" ", (WIDTH)-len(suite.Name)-len(columns))
		fmt.Printf("%s%s%s\n", suite.Name, spacing, columns)

		prevFailed := false
		for i, testCase := range suite.Cases {
			testPath := path.Join("test/cases", suite.Name, testCase.Name)
			if suite.Name == "Top Level" {
				testPath = path.Join("test/cases", testCase.Name)
			}

			tc := &suite.Cases[i]

			expected := executeTest(tf.Reference, testPath)
			target := executeTest(tf.Target, testPath)
			tc.Expected = &expected
			tc.Actual = &target
			tc.Percent = float64(expected.Duration.Nanoseconds()) / float64(target.Duration.Nanoseconds()) * 100

			prevFailed = tc.PrintResult(prevFailed)

			tf.Total++
			tf.Percent += tc.Percent
			if prevFailed {
				tf.Failed = append(tf.Failed, tc)
			}
		}
	}

	tf.Percent /= float64(tf.Total)
}

func executeTest(executable, test string) TestResult {
	command := strings.Fields(executable)
	command = append(command, test)
	cmd := exec.Command(command[0], command[1:]...)
	stdout := strings.Builder{}
	stderr := strings.Builder{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "execution error: %v", err)
		}
	}

	return TestResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}
}

/* These compare and print the test results.
 * If there is a difference in the output or error output, it will print them
 * side-by-side based on the WIDTH.
 * It also prints out how long each version took to execute and the difference
 * in how long the tested implementation took to run the same test.
 */
var divider = strings.Repeat("-", WIDTH)
var headerSpacing = strings.Repeat(" ", (WIDTH/2)-len("Expected stdout"))

// Creates the summary line and whether the result differes
func (tc TestCase) summaryVars() (string, bool) {
	succeeded := tc.Expected.ExitCode == tc.Actual.ExitCode &&
		tc.Expected.Stdout == tc.Actual.Stdout &&
		(tc.Expected.Stderr == tc.Actual.Stderr || *noFailStderr)

	result := color.GreenString("passed")
	if !succeeded {
		result = color.RedString("failed")
	}

	timing := fmt.Sprintf("%12s %12s %7.2f%%", tc.Expected.Duration, tc.Actual.Duration, tc.Percent)

	// Spacing works because len("passed") == len("failed")
	resultSpacing := strings.Repeat(" ", WIDTH-len("  [passed] ")-len(tc.Name)-len(timing))

	summary := fmt.Sprintf("  [%s] %s%s%s", result, tc.Name, resultSpacing, timing)
	return summary, !succeeded
}

func (tc TestCase) PrintResult(prevFailed bool) bool {
	summary, failed := tc.summaryVars()

	if failed && !prevFailed {
		// Don't print the divider twice for two errors in a row
		fmt.Println(divider)
	}
	fmt.Println(summary)

	if tc.Expected.ExitCode != tc.Actual.ExitCode {
		fmt.Printf("Expected exit code %d, but got %d\n", tc.Expected.ExitCode, tc.Actual.ExitCode)
	}
	if tc.Expected.Stdout != tc.Actual.Stdout {
		fmt.Printf("Expected stdout%sActual stdout\n", headerSpacing)
		printDiff(tc.Expected.Stdout, tc.Actual.Stdout)
	}
	if !*noFailStderr && tc.Expected.Stderr != tc.Actual.Stderr {
		fmt.Printf("Expected stderr%sActual stderr\n", headerSpacing)
		printDiff(tc.Expected.Stderr, tc.Actual.Stderr)
	}

	if failed {
		fmt.Println(divider)
	}
	return failed
}

func printDiff(expected, actual string) {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	for i := 0; i < len(expectedLines) && i < len(actualLines); i++ {
		spaces := (WIDTH / 2) - len(expectedLines[i])
		if spaces < 0 {
			spaces = 2
		}
		spacing := strings.Repeat(" ", spaces)
		fmt.Printf("%s%s%s\n", expectedLines[i], spacing, actualLines[i])
	}
}

func (tf TestFramework) PrintSummary() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", WIDTH))

	fmt.Println("Test summary")
	fmt.Printf("Tests run: %d\n", tf.Total)
	fmt.Printf("Succeeded: %d\n", tf.Total-len(tf.Failed))
	fmt.Printf("Failed:    %d\n", len(tf.Failed))
	fmt.Printf("Average comparative runtime: %7.2f%%\n", tf.Percent)

	fmt.Println()
	fmt.Println("Failed tests:")
	for _, tc := range tf.Failed {
		fmt.Printf("  %s\n", tc.Name)
	}
}
