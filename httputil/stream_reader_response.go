package httputil

import (
  "bytes"
  "fmt"
  "strings"
  "log"
  "net/http"
  "bufio"
  "encoding/json"
)

func renderResponseStream(w http.ResponseWriter, apiResp *http.Response, flusher http.Flusher) {
  // https://leidnedya.github.io/markweb/#https://daringfireball.net/projects/markdown/syntax#philosophy

	scanner := bufio.NewScanner(apiResp.Body)

  // tagStack := make([]string, 0, 20)

	for scanner.Scan() {
		line := scanner.Text()

    var partialResponse ReaderResponse
    json.Unmarshal([]byte(line), &partialResponse)
    text := partialResponse.Response
		fmt.Println("Received:", text)

    if text == "\n" {
      fmt.Fprintf(w, "</p><p>")
      flusher.Flush()
    } else {
      fmt.Fprintf(w, strings.ReplaceAll(text, "%", "%%")) // escape '%' in backtick string
      flusher.Flush()
    }

	}
}

func StreamReaderResponse(html string, w http.ResponseWriter) {
  payload := ReaderPayload{
    Model: "reader-lm:0.5b",
    Stream: true,
    Prompt: html,
  }

  jsonData, err := json.Marshal(payload)
  if err != nil {
    log.Println("Unable to json-stringify reader payload.")
    panic(err)
  }

	flusher, ok := w.(http.Flusher)
  if !ok {
    http.Error(w, "Streaming not supported", http.StatusInternalServerError)
    return
  }

	// Set headers to prevent buffering
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	flusher.Flush()

  fmt.Fprintf(w, strings.ReplaceAll(BEFORE_HTML, "%", "%%")) // escape '%' in backtick string
  flusher.Flush()

  resp, err := http.Post(READER_LM_URL, "application/json", bytes.NewBuffer(jsonData))
  if err != nil {
    log.Println("Unable to make a request to the reader model.")
    panic(err)
  }
	defer resp.Body.Close()

  renderResponseStream(w, resp, flusher)

  fmt.Fprintf(w, AFTER_HTML)

}
