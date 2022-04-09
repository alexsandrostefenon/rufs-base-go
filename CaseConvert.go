package rufsBase

import (
	"strings"
)

func UnderscoreToCamel(str string, isFirstUpper bool) string {
	names := strings.Split(strings.ToLower(str), "_")
	ret := ""

	for index, name := range names {
		if len(name) == 0 {
			continue
		}

		var firstChar string

		if index > 0 || isFirstUpper {
			firstChar = strings.ToUpper(name[0:1])
		} else {
			firstChar = name[0:1]
		}

		if ret += firstChar; len(name) > 1 {
			ret += name[1:]
		}
	}

	return ret
}
