package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const WIDTH = 120

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
