package main

import (
	"io/fs"
	"os"
	"path"
)

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
