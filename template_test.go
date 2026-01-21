package main

import (
	"reflect"
	"testing"
)

func TestExtractTemplateVars(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "single variable",
			command:  "curl -O {{url}}",
			expected: []string{"url"},
		},
		{
			name:     "multiple variables",
			command:  "ssh {{user}}@{{host}}",
			expected: []string{"user", "host"},
		},
		{
			name:     "duplicate variables",
			command:  "echo {{text}} and {{text}} again",
			expected: []string{"text"},
		},
		{
			name:     "no variables",
			command:  "docker ps -a",
			expected: []string{},
		},
		{
			name:     "variable with spaces",
			command:  "git clone {{ repo }} {{destination}}",
			expected: []string{"repo", "destination"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTemplateVars(tt.command)
			// Handle nil vs empty slice comparison
			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractTemplateVars(%q) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestSubstituteTemplateVars(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		values   map[string]string
		expected string
	}{
		{
			name:    "single substitution",
			command: "curl -O {{url}}",
			values: map[string]string{
				"url": "https://example.com/file.tar.gz",
			},
			expected: "curl -O https://example.com/file.tar.gz",
		},
		{
			name:    "multiple substitutions",
			command: "ssh {{user}}@{{host}}",
			values: map[string]string{
				"user": "admin",
				"host": "server.example.com",
			},
			expected: "ssh admin@server.example.com",
		},
		{
			name:    "duplicate variables",
			command: "echo {{text}} and {{text}} again",
			values: map[string]string{
				"text": "hello",
			},
			expected: "echo hello and hello again",
		},
		{
			name:     "no substitutions",
			command:  "docker ps -a",
			values:   map[string]string{},
			expected: "docker ps -a",
		},
		{
			name:    "variables with spaces",
			command: "git clone {{ repo }} {{ destination }}",
			values: map[string]string{
				"repo":        "https://github.com/user/repo.git",
				"destination": "/tmp/myrepo",
			},
			expected: "git clone https://github.com/user/repo.git /tmp/myrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubstituteTemplateVars(tt.command, tt.values)
			if result != tt.expected {
				t.Errorf("SubstituteTemplateVars(%q, %v) = %q, want %q", tt.command, tt.values, result, tt.expected)
			}
		})
	}
}
