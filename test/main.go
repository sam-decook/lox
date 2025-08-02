package main

import (
	"slices"
	"strings"
	"time"
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
