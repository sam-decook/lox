package main

import (
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
}

func main() {
	tf := TestFramework{
		Reference: "test/official-clox",
		Target:    "clox/clox_interpreter",
	}

	tf.collectSuites("test/cases")
	slices.SortFunc(tf.Suites, func(a, b *TestSuite) int {
		return strings.Compare(a.Name, b.Name)
	})

	tf.executeTests()
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
	prevFailed := false
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

		for i, testCase := range suite.Cases {
			testPath := path.Join("test/cases", suite.Name, testCase.Name)
			if suite.Name == "Top Level" {
				testPath = path.Join("test/cases", testCase.Name)
			}

			expected := executeTest(tf.Reference, testPath)
			suite.Cases[i].Expected = &expected

			target := executeTest(tf.Target, testPath)
			suite.Cases[i].Actual = &target

			prevFailed = suite.Cases[i].PrintResult(prevFailed)
		}
	}
}

func executeTest(executable, test string) TestResult {
	cmd := exec.Command(executable, test)
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

func (tc TestCase) PrintResult(prevFailed bool) bool {
	perc := float64(tc.Expected.Duration.Nanoseconds()) / float64(tc.Actual.Duration.Nanoseconds()) * 100
	timing := fmt.Sprintf("%12s %12s %7.2f%%", tc.Expected.Duration, tc.Actual.Duration, perc)

	// Spacing works because len("passed") == len("failed")
	spacing := strings.Repeat(" ", WIDTH-len("  [passed] ")-len(tc.Name)-len(timing))

	if tc.Expected.ExitCode == tc.Actual.ExitCode &&
		tc.Expected.Stdout == tc.Actual.Stdout &&
		tc.Expected.Stderr == tc.Actual.Stderr {
		fmt.Printf("  [%s] %s%s%s\n", color.GreenString("passed"), tc.Name, spacing, timing)
		return false
	}

	if !prevFailed {
		fmt.Println(divider)
	}
	fmt.Printf("  [%s] %s%s%s\n", color.RedString("failed"), tc.Name, spacing, timing)

	errorSpacing := strings.Repeat(" ", (WIDTH/2)-len("Expected stdout"))
	if tc.Expected.ExitCode != tc.Actual.ExitCode {
		fmt.Printf("Expected exit code %d, but got %d\n", tc.Expected.ExitCode, tc.Actual.ExitCode)
	} else if tc.Expected.Stdout != tc.Actual.Stdout {
		fmt.Printf("Expected stdout%sActual stdout\n", errorSpacing)
		printDiff(tc.Expected.Stdout, tc.Actual.Stdout)
	} else if tc.Expected.Stderr != tc.Actual.Stderr {
		fmt.Printf("Expected stderr%sActual stderr\n", errorSpacing)
		printDiff(tc.Expected.Stderr, tc.Actual.Stderr)
	}

	fmt.Println(divider)
	return true
}

func printDiff(expected, actual string) {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	for i := 0; i < len(expectedLines) && i < len(actualLines); i++ {
		spacing := strings.Repeat(" ", (WIDTH/2)-len(expectedLines[i]))
		fmt.Printf("%s%s%s\n", expectedLines[i], spacing, actualLines[i])
	}
}
