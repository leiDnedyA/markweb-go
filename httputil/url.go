package httputil

func hasScheme(u string) bool {
	return urlHasPrefix(u, "http:/") || urlHasPrefix(u, "https:/")
}

func urlHasPrefix(u, prefix string) bool {
	return len(u) >= len(prefix) && u[:len(prefix)] == prefix
}

