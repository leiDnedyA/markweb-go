package httputil

import (
  "bytes"
  "log"
  "net/http"
  "io"
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

