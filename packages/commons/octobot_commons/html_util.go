package octobot_commons

import (
    "fmt"
    "strings"
)

const (
    DefaultElementTextMaxSize        = 100
    MaxRecursiveExceptionCausesDepth = 20
)

type SummaryEntry struct {
    Tag     string
    Content string
}

func SummarizePageContent(htmlContent string, maxElementTextSize int) []SummaryEntry {
    cleaned := stripTags(htmlContent)
    cleaned = strings.TrimSpace(cleaned)
    if len(cleaned) > maxElementTextSize {
        cleaned = cleaned[:maxElementTextSize] + "[...]"
    }
    if cleaned == "" {
        return []SummaryEntry{}
    }
    return []SummaryEntry{{Tag: "message", Content: cleaned}}
}

func PrettyPrintSummary(summary []SummaryEntry) string {
    parts := make([]string, 0, len(summary))
    for _, entry := range summary {
        parts = append(parts, entry.Tag+"<"+entry.Content+">")
    }
    return strings.Join(parts, "; ")
}

func IsHTMLContent(htmlContent string) bool {
    return strings.Contains(htmlContent, "</html>")
}

func GetHTMLSummaryIfRelevant(htmlContent any, maxElementTextSize int) any {
    str := fmt.Sprint(htmlContent)
    if IsHTMLContent(str) {
        return PrettyPrintSummary(SummarizePageContent(str, maxElementTextSize))
    }
    return str
}

func SummarizeExceptionHTMLCauseIfRelevant(_ error, _ int) {
    // Go errors are immutable; no-op placeholder for parity.
}

func stripTags(input string) string {
    var b strings.Builder
    inTag := false
    for _, r := range input {
        switch r {
        case '<':
            inTag = true
        case '>':
            inTag = false
        default:
            if !inTag {
                b.WriteRune(r)
            }
        }
    }
    return b.String()
}
