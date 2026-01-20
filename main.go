/*

 (1) Get all links from Sitemap.xml page
	- Get xml data
	- Parse url -> <loc> tags

 (2) For each link in links visit page and scrape

 (3) Save findings to csv

*/

package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
)

func main() {

	const website = "https://docs.python-zeep.org/sitemap.xml"
	// Array containing all the known URLs in a sitemap
	knownUrls := []string{}

	// Create a Collector
	c := colly.NewCollector(colly.AllowedDomains("docs.python-zeep.org"))

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		knownUrls = append(knownUrls, e.Text)
	})

	// Start the collector
	c.Visit(website)

	file, err := os.Create("output.csv")
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Write header
	err = writer.Write([]string{"Domain", "URL"})
	if err != nil {
		log.Fatal("Error writing header:", err)
	}

	fmt.Println("All known URLs:")
	for _, url := range knownUrls {
		fmt.Println("\t", url)
		err = writer.Write([]string{"docs.python-zeep.org", url})
		if err != nil {
			log.Fatal("Error writing record:", err)
		}
	}

	fmt.Println("Collected", len(knownUrls), "URLs")

	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Error flushing data to file:", err)
	}

	fmt.Println("Successfully created output.csv")

}
