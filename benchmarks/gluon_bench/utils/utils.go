package utils

import (
	"bufio"
	"os"
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
