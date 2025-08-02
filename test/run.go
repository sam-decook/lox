package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func (tf *TestFramework) executeTests() {
	prevFailed := false
	first := true

	for _, suite := range tf.Suites {
		// if suite.Name == "benchmark" {
		// 	continue
		// 	// The benchmarks print how long they take, so they will always fail to have
		// 	// the same output
		// }

		if first {
			first = false
		} else {
			fmt.Println()
		}

		// Width of 9 for percent to take into account the '%'
		columns := fmt.Sprintf("Run time: %12s %12s %8s", "reference", "actual", "percent")
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
