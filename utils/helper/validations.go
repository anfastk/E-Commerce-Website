package helper

import "regexp"

func IsValidFullName(name string) bool {
	regex := `^[A-Za-z]{1,}(\s[A-Za-z]{1,}){1,4}$`

	re := regexp.MustCompile(regex)

	return re.MatchString(name)
}
