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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseData URLSet
	err = xml.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		log.Fatalf("Error unmarshalling XML: %v", err)
	}

	fmt.Printf("Parsed XML response.\n")

	for _, url := range responseData.URLs {
		fmt.Println("URL: ",url.Loc)
	}

	// for _, itme := range responseData.Urlset	{
	// 	fmt.Printf("Item Name: %s, Value: %d\n", item.Name, item.Value)
	// }
}