package main

import (
  "fmt"
  "log"
  "net/http"
  // "io/ioutil"
)

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
  fmt.Fprintf(w, "Hello, world!\n" + targetUrl)
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
