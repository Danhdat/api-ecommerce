package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateSlug creates SEO-friendly slug from Vietnamese text
func GenerateSlug(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove Vietnamese accents
	text = removeVietnameseAccents(text)

	// Replace spaces and special chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	text = reg.ReplaceAllString(text, "-")

	// Remove leading/trailing hyphens
	text = strings.Trim(text, "-")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	text = reg.ReplaceAllString(text, "-")

	return text
}

// removeVietnameseAccents removes Vietnamese diacritics
func removeVietnameseAccents(text string) string {
	// Vietnamese character mapping
	replacements := map[rune]rune{
		'à': 'a', 'á': 'a', 'ạ': 'a', 'ả': 'a', 'ã': 'a',
		'â': 'a', 'ầ': 'a', 'ấ': 'a', 'ậ': 'a', 'ẩ': 'a', 'ẫ': 'a',
		'ă': 'a', 'ằ': 'a', 'ắ': 'a', 'ặ': 'a', 'ẳ': 'a', 'ẵ': 'a',
		'è': 'e', 'é': 'e', 'ẹ': 'e', 'ẻ': 'e', 'ẽ': 'e',
		'ê': 'e', 'ề': 'e', 'ế': 'e', 'ệ': 'e', 'ể': 'e', 'ễ': 'e',
		'ì': 'i', 'í': 'i', 'ị': 'i', 'ỉ': 'i', 'ĩ': 'i',
		'ò': 'o', 'ó': 'o', 'ọ': 'o', 'ỏ': 'o', 'õ': 'o',
		'ô': 'o', 'ồ': 'o', 'ố': 'o', 'ộ': 'o', 'ổ': 'o', 'ỗ': 'o',
		'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ợ': 'o', 'ở': 'o', 'ỡ': 'o',
		'ù': 'u', 'ú': 'u', 'ụ': 'u', 'ủ': 'u', 'ũ': 'u',
		'ư': 'u', 'ừ': 'u', 'ứ': 'u', 'ự': 'u', 'ử': 'u', 'ữ': 'u',
		'ỳ': 'y', 'ý': 'y', 'ỵ': 'y', 'ỷ': 'y', 'ỹ': 'y',
		'đ': 'd',
	}

	var result strings.Builder
	for _, r := range text {
		if replacement, exists := replacements[r]; exists {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// NormalizeText normalizes Unicode text (fallback method)
func NormalizeText(text string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, text)
	return result
}

// ValidateSlug checks if slug is valid format
func ValidateSlug(slug string) bool {
	if slug == "" {
		return false
	}

	// Check if slug contains only lowercase letters, numbers, and hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return matched
}

// SanitizeString removes HTML tags and dangerous characters
func SanitizeString(input string) string {
	// Remove HTML tags
	htmlTag := regexp.MustCompile(`<[^>]*>`)
	input = htmlTag.ReplaceAllString(input, "")

	// Remove script tags and content
	script := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = script.ReplaceAllString(input, "")

	// Remove style tags and content
	style := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	input = style.ReplaceAllString(input, "")

	// Remove dangerous characters that could be used for XSS
	dangerous := regexp.MustCompile(`[<>\"'&]`)
	input = dangerous.ReplaceAllStringFunc(input, func(s string) string {
		switch s {
		case "<":
			return "&lt;"
		case ">":
			return "&gt;"
		case "\"":
			return "&quot;"
		case "'":
			return "&#39;"
		case "&":
			return "&amp;"
		default:
			return s
		}
	})

	return strings.TrimSpace(input)
}

// TruncateText truncates text to specified length with ellipsis
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// Find last space before maxLength
	truncated := text[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")

	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}
