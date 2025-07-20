package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	filepath.WalkDir("cases", print)
}

func print(path string, d fs.DirEntry, err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing '%q': %v\n", path, err)
	}
	entry := "File"
	if d.IsDir() {
		entry = "Directory"
	}
	fmt.Printf("%s '%s'\n", entry, d.Name())
	return nil
}
