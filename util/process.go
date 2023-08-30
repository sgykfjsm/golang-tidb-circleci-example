package util

import (
	"strings"
)

func ParseQueryLine(text string, queries []string) []string {
	if strings.HasPrefix(text, "--") || len(queries) > 0 && strings.Contains(queries[len(queries)-1], "--") {
		queries = append(queries, text)
		return queries
	}

	lines := strings.Split(text, ";")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if i == 0 && len(queries) > 0 && !strings.HasSuffix(queries[len(queries)-1], ";") || strings.HasPrefix(line, "--") {
			queries[len(queries)-1] += " " + line
		} else {
			queries = append(queries, line)
		}

		if i != len(lines)-1 {
			queries[len(queries)-1] += ";"
		}
	}

	return queries
}
