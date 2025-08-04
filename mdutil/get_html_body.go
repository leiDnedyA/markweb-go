package mdutil

import (
  "strings"
)

func GetBodyContent(html string) string {
	lowerHTML := strings.ToLower(html)
	
	startTag := "<body>"
	endTag := "</body>"

	startIndex := strings.Index(lowerHTML, startTag)
	if startIndex == -1 {
		return html
	}

	endIndex := strings.LastIndex(lowerHTML, endTag)
	if endIndex == -1 {
    return html
	}

	// Adjust startIndex to be after the start tag
	startIndex += len(startTag)

	// Ensure end tag is after start tag
	if endIndex < startIndex {
		return html
	}

	// Extract the substring from the original htmlContent
	return html[startIndex:endIndex]
}
