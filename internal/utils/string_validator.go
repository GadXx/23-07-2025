package utils

import "net/http"

func IsStringValid(str string) bool {
	head, err := http.Head(str)
	if err != nil {
		return false
	}
	switch head.Header["Content-Type"][0] {
	case "image/jpeg":
	case "image/png":
	case "application/pdf":
	default:
		return false
	}
	return true
}
