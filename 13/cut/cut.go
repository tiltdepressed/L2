// Package cut
package cut

import (
	"13/parser"
	"strings"
)

func Cut(line string, opt *parser.CutOptions) (out string, ok bool) {
	if !strings.Contains(line, opt.Delimeter) {
		if opt.Separated && !strings.Contains(line, opt.Delimeter) {
			return "", false
		}
		return line, true
	}

	parts := strings.Split(line, opt.Delimeter)

	var sb strings.Builder
	first := true

	for i, field := range parts {
		if _, selected := opt.Fields[i+1]; selected {
			if !first {
				sb.WriteString(opt.Delimeter)
			}
			sb.WriteString(field)
			first = false
		}
	}

	return sb.String(), true
}
