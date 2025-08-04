package httputil

import (
  "bytes"
  "fmt"
  "strings"
  "log"
  "net/http"
  "bufio"
  "io"
  "encoding/json"
)

type ApiRespScanner struct {
  data string
  i int
  resp *http.Response
  internalScanner *bufio.Scanner
  tokens []rune
  token string
}

func (ars *ApiRespScanner) NextToken() (string, error) {
  if ars.internalScanner == nil {
    ars.internalScanner = bufio.NewScanner(ars.resp.Body)
  }

  if ars.internalScanner.Scan() {
		line := ars.internalScanner.Text()
    var partialResponse ReaderResponse
    json.Unmarshal([]byte(line), &partialResponse)
    text := partialResponse.Response
    ars.data += text
    ars.tokens = []rune(ars.data)
  }

  ars.i++
  if (ars.i >= len(ars.tokens)) {
    log.Println(ars.data)
    return "", io.EOF
  }

  ars.token = string(ars.tokens[ars.i])
  return ars.token, nil
}

func renderResponseStream(w http.ResponseWriter, apiResp *http.Response, flusher http.Flusher) {
  scanner := &ApiRespScanner{resp: apiResp}

  tagStack := make([]string, 0, 20)

  ch := ""
  var err error
	for err == nil {
    ch, err = scanner.NextToken()
    isNewline := false

    // ignore any whitespace immediately following a newline
    if ch == "\n" {
      isNewline = true
      ch, err = scanner.NextToken()
      for err != nil && (ch == " " || ch == "\t") {
        ch, err = scanner.NextToken()
      }
      if err != nil {
        break
      }
    } 

    if ch == "\n" {
      if len(tagStack) > 0 {
        tag := tagStack[len(tagStack) - 1]
        fmt.Fprintf(w, "</%s>", tag)
        flusher.Flush()
        tagStack = tagStack [:len(tagStack) - 1]
      }
    } else if ch == "\\" { // escape character
      ch, err = scanner.NextToken()
      if err != nil {
        continue
      }
      fmt.Fprintf(w, ch)
      flusher.Flush()
    } else if isNewline && (ch == "=" || ch == "-") {
      for ch == "=" || ch == "-" && err != nil {
        ch, err = scanner.NextToken()
      }
      fmt.Fprintf(w, "\n\t\t<br>\n")
      flusher.Flush()
    } else if ch == "[" {
      if len(tagStack) == 0 {
        // If there's no current tag, put this in a p tag
        tagStack = append(tagStack, "p")
        fmt.Fprintf(w, "<p>")
        flusher.Flush()
      }
      labelText := ""
      href := ""
      finishedEarly := false

      // parse label section
      for true {
        ch, err = scanner.NextToken()
        if err != nil {
          finishedEarly = true
          break
        }
        if ch == "]" {
          break
        }
        labelText += ch
      }
      if finishedEarly {
        fmt.Fprintf(w, strings.ReplaceAll("[" + labelText, "%", "%%"))
        flusher.Flush()
        continue
      }

      // parse text section
      ch, err = scanner.NextToken()
      if ch != "(" {
        fmt.Fprintf(w, strings.ReplaceAll("[" + labelText + "]", "%", "%%"))
        flusher.Flush()
        continue
      }
      for true {
        ch, err = scanner.NextToken()
        if err != nil {
          finishedEarly = true
          break
        }
        if ch == ")" {
          break
        }
        href += ch
      }
      if finishedEarly {
        fmt.Fprintf(w, strings.ReplaceAll("[" + labelText + "](" + href, "%", "%%"))
        flusher.Flush()
        continue
      }

      // happy path
      fmt.Fprintf(w, strings.ReplaceAll("<a href=\"" + href + "\">" + labelText + "</a>", "%", "%%"))
      flusher.Flush()
    } else if ch == "!" {
      ch, err = scanner.NextToken()

      // ignore if not followed by angle bracket
      if ch != "[" {
        fmt.Fprintf(w, strings.ReplaceAll("!" + ch, "%", "%%"))
        flusher.Flush()
        continue
      }

      altText := ""
      src := ""
      finishedEarly := false

      // parse alt text section
      for true {
        ch, err = scanner.NextToken()
        if err != nil {
          finishedEarly = true
          break
        }
        if ch == "]" {
          break
        }
        altText += ch
      }
      if finishedEarly {
        fmt.Fprintf(w, strings.ReplaceAll("[" + altText, "%", "%%"))
        flusher.Flush()
        continue
      }

      // parse text section
      ch, err = scanner.NextToken()
      if ch != "(" {
        fmt.Fprintf(w, strings.ReplaceAll("[" + altText + "]", "%", "%%"))
        flusher.Flush()
        continue
      }
      for true {
        ch, err = scanner.NextToken()
        if err != nil {
          finishedEarly = true
          break
        }
        if ch == ")" {
          break
        }
        src += ch
      }
      if finishedEarly {
        fmt.Fprintf(w, strings.ReplaceAll("[" + altText + "](" + src, "%", "%%"))
        flusher.Flush()
        continue
      }

      // happy path
      fmt.Fprintf(w, strings.ReplaceAll("<img src=\"" + src + "\" alt=\"" + altText + "\">", "%", "%%"))
      flusher.Flush()
    } else {
      if len(tagStack) == 0 {
        tagStack = append(tagStack, "p")
        fmt.Fprintf(w, "<p>\n") // escape '%' in backtick string
        flusher.Flush()
      }
      fmt.Fprintf(w, strings.ReplaceAll(ch, "%", "%%")) // escape '%' in backtick string
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
