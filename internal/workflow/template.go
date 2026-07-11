package workflow

import (
	"fmt"
	"strings"

	"regexp"
)

var templatePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

func ResolveTemplate(template string, wfCtx *Context) (string, error) {
	if wfCtx == nil {
		return "", fmt.Errorf("workflow context is nil")
	}
	if strings.TrimSpace(template) == "" {
		return "", fmt.Errorf("input template is empty")
	}
	matches, err := templateMatches(template)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	last := 0
	for _, match := range matches {
		result.WriteString(template[last:match[0]])
		expression := strings.TrimSpace(template[match[2]:match[3]])
		value, err := resolveExpression(expression, wfCtx)
		if err != nil {
			return "", err
		}
		result.WriteString(value)
		last = match[1]
	}
	result.WriteString(template[last:])
	return result.String(), nil
}

func ValidateTemplate(template string, availableResults map[string]struct{}) error {
	matches, err := templateMatches(template)
	if err != nil {
		return err
	}
	for _, match := range matches {
		expression := strings.TrimSpace(template[match[2]:match[3]])
		if expression == "question" {
			continue
		}
		const resultPrefix = "results."
		if !strings.HasPrefix(expression, resultPrefix) {
			return fmt.Errorf("unknown input template expression %q", expression)
		}
		key := strings.TrimSpace(strings.TrimPrefix(expression, resultPrefix))
		if key == "" {
			return fmt.Errorf("result template key is empty")
		}
		if _, ok := availableResults[key]; !ok {
			return fmt.Errorf("workflow result %q is not available before this node", key)
		}
	}
	return nil
}

func templateMatches(template string) ([][]int, error) {
	if strings.TrimSpace(template) == "" {
		return nil, fmt.Errorf("input template is empty")
	}
	matches := templatePattern.FindAllStringSubmatchIndex(template, -1)
	if strings.Count(template, "{{") != len(matches) || strings.Count(template, "}}") != len(matches) {
		return nil, fmt.Errorf("invalid input template %q", template)
	}
	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}

func resolveExpression(expression string, wfCtx *Context) (string, error) {
	if expression == "question" {
		return wfCtx.Question, nil
	}
	const resultPrefix = "results."
	if !strings.HasPrefix(expression, resultPrefix) {
		return "", fmt.Errorf("unknown input template expression %q", expression)
	}
	key := strings.TrimSpace(strings.TrimPrefix(expression, resultPrefix))
	if key == "" {
		return "", fmt.Errorf("result template key is empty")
	}
	value, ok := wfCtx.Results[key]
	if !ok {
		return "", fmt.Errorf("workflow result %q is not available", key)
	}
	return value, nil
}
