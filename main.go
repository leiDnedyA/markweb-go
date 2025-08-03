package main

import (
  "leidnedya/markweb/httputil"
  "leidnedya/markweb/mdutil"
  "fmt"
  "log"
  "net/http"
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

  log.Println("Getting HTML for " + targetUrl)
  htmlContent, err := httputil.GetPageHTML(targetUrl)
  if err != nil {
    fmt.Println("Error fetching HTML.")
    return
  }

  log.Println("Converting page " + targetUrl + " to markdown")
  mdContent, err := httputil.HTMLToMD(htmlContent)
  if err != nil {
    fmt.Println("Error converting page to markdown.")
    return
  }

  log.Println("Parsing markdown to simplified HTML")
  cleanHTML, err := mdutil.MarkdownToHTML(mdContent)
  if err != nil {
    fmt.Println("Error parsing markdown to HTML")
    return
  }

  r.Header.Add("Content-Type", "text/htm")
  fmt.Fprintf(w, cleanHTML)
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
