package httputil

import (
  "strings"
  "log"
  "net/http"
  "io"
  "net/url"
)


func GetPageHTML(rawURL string) (string, error) {
  if !hasScheme(rawURL) {
    rawURL = "http://" + rawURL
  } else {
    // The scheme gets converted from :// to :/ up when the request is made
    rawURL = strings.Replace(rawURL, ":/", "://", 1)
  }

  u, err := url.Parse(rawURL)
  if err != nil {
    log.Println("Invalid URL: " + rawURL)
    return "", err
  }

  resp, err := http.Get(u.String())
  if err != nil {
    log.Println("Failed to request page: " + u.String())
    return "", err
  }

  defer resp.Body.Close()

  bodyBytes, err := io.ReadAll(resp.Body)
  if err != nil {
    log.Println("Failed to parse response body for page: " + u.String())
    return "", err
  }

  bodyString := string(bodyBytes)
  return bodyString, nil
}
