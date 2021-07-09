package yaml

import (
	"strings"
)

// DeriveSeqIndentStyle derives the sequence indentation annotation value for the resource,
// originalYAML is the input yaml string,
// the style is decided by deriving the existing sequence indentation of first sequence node
func DeriveSeqIndentStyle(originalYAML string) string {
	lines := strings.Split(originalYAML, "\n")
	for i, line := range lines {
		elems := strings.SplitN(line, "- ", 2)
		if len(elems) != 2 {
			continue
		}
		// prefix of "- " must be sequence of spaces
		if strings.Trim(elems[0], " ") != "" {
			continue
		}
		numSpacesBeforeSeqElem := len(elems[0])

		if i == 0 {
			continue
		}

		// keyLine is the line before the first sequence element
		keyLine := lines[i-1]

		numSpacesBeforeKeyElem := len(keyLine) - len(strings.TrimLeft(keyLine, " "))
		trimmedKeyLine := strings.Trim(keyLine, " ")
		if strings.HasSuffix(trimmedKeyLine, "|") || strings.HasSuffix(trimmedKeyLine, "|-") {
			// this is not a sequence node, it is a wrapped sequence node string
			// ignore it
			continue
		}

		if numSpacesBeforeSeqElem == numSpacesBeforeKeyElem {
			return string(CompactSeqIndent)
		}

		if numSpacesBeforeSeqElem-numSpacesBeforeKeyElem == 2 {
			return string(WideSeqIndent)
		}
	}

	return string(CompactSeqIndent)
}
