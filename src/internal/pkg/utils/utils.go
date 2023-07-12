package utils

func ArrayContainsStr(haystack []string, needle string) bool {
	has := false
	for i := 0; i < len(haystack); i++ {
		if haystack[i] == needle {
			has = true
			break
		}
	}
	return has
}
