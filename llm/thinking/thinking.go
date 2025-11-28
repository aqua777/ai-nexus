package thinking

import (
	"regexp"
	"strings"
)

const (
	thinkingTagEnd = "</think>"
	emptyStr       = ""
)

var (
	// thinkingRegexOld = regexp.MustCompile(`(?s)^(.*)<think>(.*)</think>(.*)$`)
	thinkingRegex = regexp.MustCompile(`(?s)<think>(.*?)</think>`)
)

// ProcessContent separates thinking from response based on two patterns:
// 1. No opening tag: from beginning until </think>, remainder is response
// 2. With tags: all content between <think> and </think> tags concatenated, remainder is response
func ProcessContent(content string) (response, thinking string) {
	// Pattern 2: Handle content between <think> and </think> tags
	matches := thinkingRegex.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		// Extract all thinking parts and concatenate them
		var thinkingParts []string
		for _, match := range matches {
			if len(match) > 1 {
				thinkingParts = append(thinkingParts, strings.TrimSpace(match[1]))
			}
		}
		thinking = strings.Join(thinkingParts, " ")

		// Remove all <think>...</think> blocks to get the response
		response = thinkingRegex.ReplaceAllString(content, "")
		response = strings.TrimSpace(response)
		return response, thinking
	}

	// Pattern 1: No opening tag, from beginning until </think>
	if strings.Contains(content, thinkingTagEnd) {
		parts := strings.SplitN(content, thinkingTagEnd, 2)
		if len(parts) == 2 {
			thinking = strings.TrimSpace(parts[0])
			response = strings.TrimSpace(parts[1])
			return response, thinking
		}
	}

	// No thinking tags found, entire content is response
	return content, ""
}