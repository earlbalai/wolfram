package main

import (
 "encoding/xml"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
)

const (
  wolfram_api_key = "API-KEY-HERE" // Sign up for an API key here: http://products.wolframalpha.com/developers/
  wolfram_api_url = "http://api.wolframalpha.com"
)

// Wolfram|Alpha Structs
type QueryResult struct {
  success bool `xml:"success,attr"`
  XMLName       xml.Name `xml:"queryresult"`
  Pods        []Pod         `xml:"pod"`
}

type Pod struct {
  XMLName xml.Name `xml:"pod"`
  Title      string `xml:"title,attr"`
  Subpods []SubPod `xml:"subpod"`
}

type SubPod struct {
  XMLName   xml.Name `xml:"subpod"`
  Title     string   `xml:"title,attr"`
  PlainText string   `xml:"plaintext"`
}

// Twilio SMS Structs
type SmsBlock struct {
  XMLName xml.Name `xml:"Response"`
  Message string   `xml:",omitempty"`
}

// Query functions
//
func web_query(w http.ResponseWriter, r *http.Request) {
  query := r.FormValue("query")
  response := "Empty query"
  if query != "" {
  response = compute(query)
} else {
  response = "No Data"
}

  w.Write([]byte(fmt.Sprintf("Wolfram Go\nData Source: Wolfram|Alpha\nVersion: 0.1a\n\nQuestion\n-----------------------\n%s\n\nAnswer\n-------------------\n%s", query, response)))

}

func sms_query(w http.ResponseWriter, r *http.Request) {
  sender := r.FormValue("From")
  message := r.FormValue("Body")
  reply := SmsBlock{Message: "There was a problem validating your request... Please try again later."}
  if message != "" {
      reply = SmsBlock{
      Message: fmt.Sprintf("\nWolfram\n-------------------------------\n%s", compute(message)),
    }
  } else {
    reply = SmsBlock { Message: "Error processing your question..." }
  }

  x, err := xml.Marshal(reply)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/xml")
  w.Write(x)

  fmt.Printf("Sender: %s Message: %d\n", sender, message)
}

func call_query(w http.ResponseWriter, r *http.Response) {
 // Handle calls
}

// End query functions

func compute(question string) string {

    var Url *url.URL
    Url, err := url.Parse(wolfram_api_url)
    if err != nil {
        panic("boom")
    }

    Url.Path += "/v2/query"
    parameters := url.Values{}
    parameters.Add("appid", wolfram_api_key)
    parameters.Add("input", question)
    parameters.Add("format", "plaintext")
    Url.RawQuery = parameters.Encode()


  res, err := http.Get(Url.String())
  if err != nil {
    panic(err.Error())
  }


  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    panic(err.Error())
  }

  dat := QueryResult{}
  _ = xml.Unmarshal(body, &dat)

  response := "Error fetching data..."

  if val,ok := dat.Pods[1]; !ok {
      response = "There was an error processing your question."
  } else {

    response = dat.Pods[1].Subpods[0].PlainText

    fmt.Printf("Body: %s \n Data: %s", string(body), response)
}

  return response
}

func main() {
  fmt.Println("Knowledge Link\nVersion: 0.1a\nReady!")
  http.HandleFunc("/ask", web_query)
  http.HandleFunc("/sms", sms_query)
  http.HandleFunc("/call", call_query)
  http.ListenAndServe(":3000", nil)
}