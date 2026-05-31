package web

import (
	"net/url"
	"strings"
)

const whatsappPhone = "998901234567"

func urlQuery(value string) string {
	return url.QueryEscape(value)
}

func whatsappURL(message string) string {
	return "https://wa.me/" + whatsappPhone + "?text=" + url.QueryEscape(message)
}

func canonical(baseURL, path string) string {
	if path == "" {
		path = "/"
	}
	return strings.TrimRight(baseURL, "/") + path
}

func queryMessage(values url.Values, key string) string {
	return values.Get(key)
}
