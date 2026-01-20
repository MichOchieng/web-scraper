/*

 (1) Get all links from Sitemap.xml page
	- Get xml data
	- Parse url -> <loc> tags

 (2) For each link in links visit page and scrape

 (3) Save findings to csv

*/

package main

import (
	"fmt"
	"github.com/gocolly/colly"
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

	fmt.Println("All known URLs:")
	for _, url := range knownUrls {
		fmt.Println("\t", url)
	}
	fmt.Println("Collected", len(knownUrls), "URLs")


}