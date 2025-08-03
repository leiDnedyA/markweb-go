package main

import (
  "bytes"
  "strings"
  "fmt"
  "log"
  "net/http"
  "io"
  "net/url"
  "encoding/json"
)

const READER_LM_URL = "http://localhost:11434/api/generate"

type ReaderPayload struct {
  Model string `json:"model"`
  Stream bool `json:"stream"`
  Prompt string `json:"prompt"`
}

type ReaderResponse struct {
  Response string `json:"response"`
}

func hasScheme(u string) bool {
	return urlHasPrefix(u, "http:/") || urlHasPrefix(u, "https:/")
}

func urlHasPrefix(u, prefix string) bool {
	return len(u) >= len(prefix) && u[:len(prefix)] == prefix
}

func getPageHTML(rawURL string) (string, error) {
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

func htmlToMD(html string) (string, error) {
  payload := ReaderPayload{
    Model: "reader-lm:1.5b",
    Stream: false,
    Prompt: html,
  }

  jsonData, err := json.Marshal(payload)
  if err != nil {
    fmt.Println("Unable to json-stringify reader payload.")
    panic(err)
  }

  resp, err := http.Post(READER_LM_URL, "application/json", bytes.NewBuffer(jsonData))
  if err != nil {
    log.Println("Request to reader API failed.")
    return "", err
  }
  defer resp.Body.Close()

  bodyBytes, err := io.ReadAll(resp.Body)
  if err != nil {
    log.Println("Unable to parse response from reader API")
    return "", err
  }

  bodyJSON := string(bodyBytes)

  var body ReaderResponse

  json.Unmarshal([]byte(bodyJSON), &body)

  markdown := body.Response

  return markdown, nil
}

func homepageHandler(w http.ResponseWriter, r *http.Request) {
    r.Header.Set("Content-Type", "text/html")
    fmt.Fprintf(w, `
    <!DOCTYPE html>
    <form id="form">
      <input type="text" id="input">
      <input type="submit">
    </form>
    <script>
      const input = document.querySelector("#input");
      const form = document.querySelector("#form");

      form.onsubmit = (e) => {
        e.preventDefault();
        window.location.pathname = '/' + input.value;
      }
    </script>
    `)
}

func readerHandler(w http.ResponseWriter, r *http.Request) {
  path := r.URL.Path
  targetUrl := path[1:]
  log.Println("Getting HTML for " + targetUrl)
  htmlContent, err := getPageHTML(targetUrl)
  if err != nil {
    fmt.Println(err)
    return
  }
  log.Println("Converting page " + taragetUrl + " to markdown")
  mdContent, err := htmlToMD(htmlContent)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Fprintf(w, mdContent)
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
  path := r.URL.Path

  if (len(path) <= 1) {
    homepageHandler(w, r)
    return
  }
  readerHandler(w, r)
  return
}

func main() {
  http.HandleFunc("/", reqHandler)
  log.Println("Starting server on :8080")
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}
