package rufsBase

import (
	"regexp"
	"strings"
)

func UnderscoreToCamel(str string, isFirstUpper bool) string {
	names := strings.Split(strings.TrimSpace(strings.ToLower(str)), "_")
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

func CamelToUnderscore(str string) string {
	str = strings.TrimSpace(str)
	re := regexp.MustCompile(`([A-Z])`)
	ret := strings.ToLower(re.ReplaceAllString(str, "_$1"))

	if strings.HasPrefix(ret, "_") {
		ret = ret[1:]
	}

	return ret
}
