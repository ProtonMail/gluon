package utils

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ReadLinesFromFile(path string) ([]string, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	lines := make([]string, 0, 16)

	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return lines, nil
}

func LoadFilesFromDirectory(path string, filter func(string, fs.FileInfo) bool) ([]string, error) {
	var files []string

	if err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !filter(path, info) {
			return nil
		}

		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files = append(files, string(bytes))

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func LoadEMLFilesFromDirectory(path string) ([]string, error) {
	return LoadFilesFromDirectory(path, func(s string, info fs.FileInfo) bool {
		return !info.IsDir() && strings.HasSuffix(s, ".eml")
	})
}
