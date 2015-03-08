package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var wolfram_api_key string

const (
	wolfram_api_url = "http://api.wolframalpha.com"
)

// Wolfram|Alpha Structs
type QueryResult struct {
	success bool     `xml:"success,attr"`
	XMLName xml.Name `xml:"queryresult"`
	Pods    []Pod    `xml:"pod"`
}

type Pod struct {
	XMLName xml.Name `xml:"pod"`
	Title   string   `xml:"title,attr"`
	SubPods []SubPod `xml:"subpod"`
}

type SubPod struct {
	XMLName   xml.Name `xml:"subpod"`
	title     string   `xml:"title,attr"`
	PlainText string   `xml:"plaintext"`
}

// Twilio SMS Structs
type smsBlock struct {
	XMLName xml.Name `xml:"Response"`
	Message string   `xml:",omitempty"`
}

// Query functions
//
func webQuery(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	response := "Empty query"
	if query != "" {
		response = compute(query)
	} else {
		response = "No Data"
	}

	w.Write([]byte(fmt.Sprintf("Data Source: Wolfram|Alpha\n\nQuestion\n-----------------------\n%s\n\nAnswer\n-------------------\n%s", query, response)))

}

func smsQuery(w http.ResponseWriter, r *http.Request) {
	sender := r.FormValue("From")
	message := r.FormValue("Body")
	reply := smsBlock{Message: "There was a problem validating your request... Please try again later."}
	if message != "" {
		reply = smsBlock{
			Message: fmt.Sprintf("\nWolfram\n-------------------------------\n%s", compute(message)),
		}
	} else {
		reply = smsBlock{Message: "Error processing your question..."}
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

func callQuery(w http.ResponseWriter, r *http.Response) {
	// Handle calls
}

// End query functions

func compute(question string) string {

	var Url *url.URL
	Url, err := url.Parse(wolfram_api_url)
	if err != nil {
		panic("bork'd")
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

	if len(dat.Pods) <= 0 {
		response = "There was an error processing your question."
	} else {

		response = dat.Pods[1].SubPods[0].PlainText

		fmt.Printf("Body: %s \n Data: %s", string(body), response)
	}

	return response
}

func main() {

	flag.StringVar(&wolfram_api_key, "key", "", "Wolfram|Alpha Developer API Key")

	flag.Parse()

	if len(wolfram_api_key) <= 0 {
		log.Fatal("Please specify the -key argument with your api key.\nExample: -key \"3KR2D7-4G3962597F\"")
	}

	fmt.Println("Ready!")
	http.HandleFunc("/ask", webQuery)
	http.HandleFunc("/sms", smsQuery)
	//http.HandleFunc("/call", callQuery)
	http.ListenAndServe(":3000", nil)
}
