/*

 (1) Get all links from Sitemap.xml page
	- Get xml data
	- Parse url -> <loc> tags

 (2) For each link in links visit page and scrape

 (3) Save findings to csv

*/

package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"log"
)


type URLSet struct {
    URLs []URL `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
	LastMod string `xml:"lastmod"`
	ChageFreq string `xml:"chagefreq"` 
	Priority string `xml:"priority"`
}

func getUrls(bodyBytes []byte) (URLSet) {
	var responseData URLSet
	err := xml.Unmarshal(bodyBytes, &responseData)

	if err != nil {
		log.Fatalf("Error unmarshalling XML: %v", err)
	}

	return responseData
}

func readResponseXML(response io.Reader) ([]byte, error) {
	bodyBytes, err := io.ReadAll(response)

	if err != nil {
		log.Fatal(err)
	}

	return bodyBytes, err
}

func main() {

	// 1: Get xml from Sitemap.xml
	const website = "https://docs.python-zeep.org/sitemap.xml"

	resp,err := http.Get(website)

	defer resp.Body.Close()

	if err != nil {
		log.Fatalf("Error making HTTP request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes,err := readResponseXML(resp.Body)

	data := getUrls(bodyBytes)

	for _, url := range data.URLs {
		fmt.Println("URL: ",url.Loc)
	}
}