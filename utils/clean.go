package utils

import (
	"regexp"
	"strings"
)

// CleanText for clean text
func CleanText(text string, maxLength int) string {
	if len(text) < 5 {
		return ""
	}

	if strings.Contains(text, "\n") {
		sections := strings.Split(text, "\n")
		newText := sections[0]
		for idx, s := range sections {
			// Append sections until we reach the max length
			if idx > 0 && len(newText) < maxLength {
				newText = newText + " " + s
			}
		}
		text = newText
	}

	var charMap = map[string]string{
		"′":          "",
		"'":          "",
		"|":          "",
		"\u0026":     "and", // ampersand
		"\u1ebd":     "e",
		"\u200b":     " ",
		"\u200e":     " ",
		"\u2010":     "-",
		"\u2013":     "-",
		"\u2014":     "-",
		"\u2018":     "",
		"\u2019":     "",
		"\u2022":     "-",
		"\u2026":     "...",
		"\u2028":     "",
		"\u2033":     "\"",
		"\u2034":     "\"",
		"\u2035":     "'",
		"\u2036":     "\"",
		"\u2037":     "\"",
		"\u2038":     ".",
		"\u2044":     "/",
		"\u201a":     ",",
		"\u201b":     "'",
		"\u201c":     "\"",
		"\u201d":     "\"",
		"\u201e":     "\"",
		"\u201f":     "\"",
		"\u2122":     "",
		"\u2600":     "",
		"\u263a":     "",
		"\u26fa":     "",
		"\u27a2":     ">",
		"\ufe0f":     "",
		"\xa0":       " ",
		"\xa2":       "",
		"\xae":       "",
		"\xbd":       "",
		"\xde":       "",
		"\xe2":       "",
		"\xe9":       "",
		"\xfc":       "u",
		"\U0001f44c": "",
		"\U0001f44d": "",
		"\U0001f642": "",
		"\U0001f601": "",
		"\U0001f690": "",
		"\U0001f334": "",
		"\U0001f3dd": "",
		"\U0001f3fd": "",
		"\U0001f3d6": "",
		"\U0001f3a3": "",
		"\U0001f525": "", // flame
		"\U0001f60a": "", // smiley
	}

	// I initially attempted to use strings.Replace but it gave intermittent results.
	// It also had to scan the entire string for each character to replace.
	// With this approach you scan the string once, replacing all characters.
	newText := ""
	for _, c := range text {
		newC, ok := charMap[string(c)]
		// If not found, use the original
		if !ok {
			newC = string(c)
		}
		newText = newText + newC
	}
	text = newText

	if len(text) > maxLength {
		return text[0:maxLength-3] + "..."
	}

	return text
}

// CleanString for clean string v2
func CleanString(param string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return ""
	}
	return reg.ReplaceAllString(param, "")
}
