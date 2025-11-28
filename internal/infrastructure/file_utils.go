package infrastructure

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func ReadLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func ExtractLanguageFromPath(path string) string {
	dir := filepath.Dir(path)
	parts := strings.Split(dir, "/")

	for _, part := range parts {
		if part == "ita" || part == "eng" {
			return part
		}
	}

	return "ita"
}

func RemoveMarkdownHeaders(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {

		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
