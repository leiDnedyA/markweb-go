package httputil

import (
  "bytes"
  "fmt"
  "strings"
  "log"
  "net/http"
  "io"
  "bufio"
  "encoding/json"
)


const READER_LM_URL = "http://localhost:11434/api/generate"

const PRE_HTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Markweb</title>
    <style>
      body {
        margin: 0;
        padding: 0;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
        background-color: #121212;
        color: #e0e0e0;
        display: flex;
        flex-direction: column;
        align-items: center;
        line-height: 1.6;
        font-size: 1rem;
        padding: 1rem;
      }
      
      #status-bar {
        position: fixed;
        top: 0;
        left: 0;
        background-color: #666;
        font-size: 1em;
        display: none;
        z-index: 999;
      }
      
      #content {
        width: 100%;
        max-width: 700px;
        padding: 1rem;
        box-sizing: border-box;
        display: flex;
        flex-direction: column;
      }
      
      h1, h2, h3, h4, h5, h6 {
        color: #ffffff;
        margin-top: 2rem;
        margin-bottom: 1rem;
        line-height: 1.3;
      }
      
      h1 { font-size: 2rem; }
      h2 { font-size: 1.75rem; }
      h3 { font-size: 1.5rem; }
      h4 { font-size: 1.25rem; }
      h5 { font-size: 1.1rem; }
      h6 { font-size: 1rem; }
      
      p {
        position: relative;
        display: inline-block;
        margin-bottom: 1rem;
        color: #cccccc;
      }
      
      p > span.tooltip {
        position: absolute;
        top: 100%;     /* Position below the button */
        left: 0;
        background: black;
        color: white;
        padding: 5px 10px;
        border-radius: 4px;
        white-space: nowrap;
        display: none;  /* Initially hidden */
      }
      
      span.bookmark-indicator {
        position: absolute;
        top: 10%;
        right: 100%;
        color: #888;
        padding: 5px 10px;
        border-radius: 4px;
        white-space: nowrap;
        display: none;
      }
      
      p:hover span.bookmark-indicator {
        display: block;
      }
      
      span.bookmarked {
        display: block;
      }
      
      p:hover .tooltip {
        display: block;
      }
      
      ul, ol {
        padding-left: 1.5rem;
        margin-bottom: 1rem;
      }
      
      li {
        margin-bottom: 0.5rem;
      }
      
      pre {
        background-color: #1e1e1e;
        color: #dcdcdc;
        padding: 1rem;
        border-radius: 8px;
        overflow-x: auto;
        font-family: 'Courier New', monospace;
      }
      
      code {
        background-color: #1e1e1e;
        color: #e0e0e0;
        padding: 0.2em 0.4em;
        border-radius: 4px;
        font-family: 'Courier New', monospace;
      }
      
      blockquote {
        border-left: 4px solid #888;
        padding-left: 1rem;
        color: #aaaaaa;
        margin: 1rem 0;
        font-style: italic;
      }
      
      a {
        color: #fff;
        text-decoration: none;
        margin-left: 3px;
        margin-right: 3px;
      }
      a:hover {
        text-decoration: underline;
      }
      
      a.new-tab {
        color: #fff;
        margin: 0;
        margin-right: 2px;
        margin-left: 5px;
        padding-left: 0px;
        padding-right: 5px;
      }
      
      hr {
        border: none;
        border-top: 1px solid #444;
        margin: 2rem 0;
      }
      
      img {
        max-width: 100%;
        height: auto;
        border-radius: 6px;
        margin: 1rem 0;
      }
      
      select {
        background-color: #1e1e1e;
        color: #e0e0e0;
        border: 1px solid #333;
        font-size: 1rem;
        cursor: pointer;
      }
      
      select:focus {
        border-color: #888;
        box-shadow: 0 0 0 2px rgba(190, 190, 190, 0.2);
        outline: none;
      }
      
      option {
        background-color: #1e1e1e;
        color: #e0e0e0;
      }
      
      #input-form, #bookmark-container {
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: center;
        gap: 5px;
        width: 100%;
        margin-bottom: 5px;
      }
      
      input[type="text"],
      textarea {
        background-color: #1e1e1e;
        color: #e0e0e0;
        border: 1px solid #333;
        border-radius: 6px;
        font-size: 1rem;
        box-sizing: border-box;
        outline: none;
      }
      
      input[type="text"]:focus,
      textarea:focus {
        border-color: #888;
        box-shadow: 0 0 0 2px rgba(190, 190, 190, 0.2);
      }
      
      input[type="submit"],
      button {
        background-color: #888;
        color: #121212;
        border: none;
        padding: 3px 5px;
        font-weight: bold;
        border-radius: 2px;
        cursor: pointer;
      }
      
      input[type="submit"]:hover,
      button:hover {
        background-color: #ccc;
      }
      
      #next-bookmark {
        position: fixed;
        bottom: 0;
        left: 0;
        display: flex;
        flex-direction: row;
        align-items: center;
        z-index: 999;
      }
      
      #page-bookmark-button {
        background: transparent;
        display: block;
        padding-top: 6px;
      }
      
      #page-bookmark-button:hover {
        background: #333;
      }
      
      @media (max-width: 768px) {
        body {
          font-size: 1.05rem;
        }
      
        #content {
          width: 100%;
          padding: 1rem;
        }
      
        h1 { font-size: 1.75rem; }
        h2 { font-size: 1.5rem; }
        h3 { font-size: 1.25rem; }
      
        span.bookmark-indicator {
          padding-right: 2px;
        }
      }
    </style>
  </head>
  <body>
    <div id="content">
      <p>
  `

type ReaderPayload struct {
  Model string `json:"model"`
  Stream bool `json:"stream"`
  Prompt string `json:"prompt"`
}

type ReaderResponse struct {
  Response string `json:"response"`
}

func HTMLToMD(html string) (string, error) {
  payload := ReaderPayload{
    Model: "reader-lm:0.5b",
    Stream: false,
    Prompt: html,
  }

  jsonData, err := json.Marshal(payload)
  if err != nil {
    log.Println("Unable to json-stringify reader payload.")
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

  resp, err := http.Post(READER_LM_URL, "application/json", bytes.NewBuffer(jsonData))
  if err != nil {
    log.Println("Unable to make a request to the reader model.")
    panic(err)
  }
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

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

  fmt.Fprintf(w, strings.ReplaceAll(PRE_HTML, "%", "%%")) // escape '%' in backtick string
  flusher.Flush()

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
      fmt.Fprintf(w, text)
      flusher.Flush()
    }
	}

  fmt.Fprintf(w, `
      </p>
    </div>
  </body>
</html>
  `)

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading stream:", err)
	}
}
