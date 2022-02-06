package util

func GetFirstNonEmptyString(strings []string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return ""
}
