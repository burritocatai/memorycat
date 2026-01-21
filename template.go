package main

import (
	"regexp"
	"strings"
)

// ExtractTemplateVars extracts all unique template variables from a command
// Template variables are in the format {{variable_name}}
func ExtractTemplateVars(command string) []string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(command, -1)

	// Use a map to track unique variables
	seen := make(map[string]bool)
	var vars []string

	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			if !seen[varName] {
				seen[varName] = true
				vars = append(vars, varName)
			}
		}
	}

	return vars
}

// SubstituteTemplateVars replaces template variables with their values
func SubstituteTemplateVars(command string, values map[string]string) string {
	result := command
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result = re.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the variable name from the match
		varName := strings.TrimSpace(match[2 : len(match)-2])
		if value, ok := values[varName]; ok {
			return value
		}
		// If no value found, return the original placeholder
		return match
	})

	return result
}
