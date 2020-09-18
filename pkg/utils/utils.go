package utils

import (
	"html"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//GetClient creates simple http client
func GetClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 260 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 260 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 300,
		Transport: netTransport,
	}
	return netClient
}

//Clean removes HTML, double spaces and other
func Clean(in string) (out string) {
	re := regexp.MustCompile(`<[^>]*>`)
	re2 := regexp.MustCompile(`\s+`)
	out = re.ReplaceAllString(in, " ")
	out = html.UnescapeString(out)
	out = strings.ReplaceAll(out, "\u00a0", " ")
	out = re2.ReplaceAllString(out, " ")
	return strings.TrimSpace(out)

}
