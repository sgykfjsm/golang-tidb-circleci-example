package util_test

import (
	"testing"

	"github.com/sgykfjsm/golang-tidb-circleci-example/util"
	"gotest.tools/v3/assert"
)

func TestParseQueryLine(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		existingQueries []string
		expected        []string
	}{
		{
			name:            "Comment line",
			input:           "-- This is a comment line",
			existingQueries: []string{},
			expected:        []string{"-- This is a comment line"},
		},
		{
			name:            "Comment line with existing queries",
			input:           "-- This is a comment line",
			existingQueries: []string{"SHOW TABLES;"},
			expected:        []string{"SHOW TABLES;", "-- This is a comment line"},
		},
		{
			name:            "Existing query is empty",
			input:           "CREATE DATABASE IF NOT EXISTS test;",
			existingQueries: []string{},
			expected:        []string{"CREATE DATABASE IF NOT EXISTS test;"},
		},
		{
			name:            "Adding to existing queries",
			input:           "USE test;",
			existingQueries: []string{"CREATE DATABASE IF NOT EXISTS test;"},
			expected:        []string{"CREATE DATABASE IF NOT EXISTS test;", "USE test;"},
		},
		{
			name:            "Adding to existing queries2",
			input:           "id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,",
			existingQueries: []string{"CREATE TABLE IF NOT EXISTS users ("},
			expected:        []string{"CREATE TABLE IF NOT EXISTS users ( id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,"},
		},
		{
			name:            "Adding to existing queries3",
			input:           "name VARCHAR(64) NOT NULL );",
			existingQueries: []string{"CREATE TABLE IF NOT EXISTS users ( id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY,"},
			expected:        []string{"CREATE TABLE IF NOT EXISTS users ( id BIGINT NOT NULL AUTO_RANDOM PRIMARY KEY, name VARCHAR(64) NOT NULL );"},
		},
		{
			name:            "Multiple queries in one line",
			input:           "USE test; CREATE TABLE example (id INT);",
			existingQueries: []string{},
			expected:        []string{"USE test;", "CREATE TABLE example (id INT);"},
		},
		{
			name:            "Multiple queries in one line with existing queries",
			input:           "USE test; CREATE TABLE example (id INT);",
			existingQueries: []string{"SHOW TABLES;"},
			expected:        []string{"SHOW TABLES;", "USE test;", "CREATE TABLE example (id INT);"},
		},
		{
			name:            "Multiple queries in one line with a comment line",
			input:           "USE test; CREATE TABLE example (id INT); -- comment",
			existingQueries: []string{},
			expected:        []string{"USE test;", "CREATE TABLE example (id INT); -- comment"},
		},
		{
			name:            "Add query to the line ending with a comment",
			input:           "name VARCHAR(32),",
			existingQueries: []string{"USE test;", "CREATE TABLE example (id INT, -- comment"},
			expected:        []string{"USE test;", "CREATE TABLE example (id INT, -- comment", "name VARCHAR(32),"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := util.ParseQueryLine(tt.input, tt.existingQueries)
			assert.DeepEqual(t, tt.expected, actual)
		})
	}
}
