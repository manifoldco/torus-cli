// +build !windows

package promptui

import (
	"fmt"
	"strconv"
	"strings"
)

func styler(attrs ...attribute) func(string) string {
	attrstrs := make([]string, len(attrs))
	for i, v := range attrs {
		attrstrs[i] = strconv.Itoa(int(v))
	}

	seq := strings.Join(attrstrs, ";")

	return func(s string) string {
		end := ""
		if !strings.HasSuffix(s, ResetCode) {
			end = ResetCode
		}
		return fmt.Sprintf("%s%sm%s%s", esc, seq, s, end)
	}
}
