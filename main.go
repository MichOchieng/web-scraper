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
	"strings"
)

type Product struct {
	URL         string
	Title       string
	Brand       string
	Description string
	Price       string
	ImageURL    string
	Category    string
	Specs       string
	CustomInfo  string
}

var products []Product

func main() {

	const website = "https://www.rkfurnituregallery.ca/sitemap.xml"
	// Array containing all the known URLs in a sitemap
	knownUrls := []string{}

	// Create a Collector
	c := colly.NewCollector(colly.AllowedDomains("www.rkfurnituregallery.ca"))

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		knownUrls = append(knownUrls, e.Text)
	})

	// Start the sitemap collector
	c.Visit(website)

	// Create product scraper collector
	productCollector := colly.NewCollector(colly.AllowedDomains("www.rkfurnituregallery.ca"))

	// Extract title from page title
	productCollector.OnHTML("title", func(e *colly.HTMLElement) {
		fmt.Printf("Processing: %s\n", e.Text)
	})

	// Extract product data from OpenGraph meta tags
	productCollector.OnHTML("meta[property='og:title']", func(e *colly.HTMLElement) {
		product := Product{URL: e.Request.URL.String()}
		product.Title = e.Attr("content")
		products = append(products, product)
	})

	// Extract description
	productCollector.OnHTML("meta[property='og:description']", func(e *colly.HTMLElement) {
		if len(products) > 0 {
			products[len(products)-1].Description = e.Attr("content")
		}
	})

	// Extract image
	productCollector.OnHTML("meta[property='og:image']", func(e *colly.HTMLElement) {
		if len(products) > 0 {
			products[len(products)-1].ImageURL = e.Attr("content")
		}
	})

	// Extract brand and category from page title
	productCollector.OnHTML("h1", func(e *colly.HTMLElement) {
		if len(products) > 0 {
			title := strings.TrimSpace(e.Text)
			// Extract brand from title (usually first word before "by")
			if strings.Contains(title, " by ") {
				parts := strings.Split(title, " by ")
				if len(parts) >= 2 {
					products[len(products)-1].Brand = strings.TrimSpace(parts[0])
					products[len(products)-1].Category = strings.TrimSpace(parts[1])
				}
			} else {
				// Fallback: try to extract from URL path
				urlParts := strings.Split(e.Request.URL.String(), "/")
				if len(urlParts) > 1 {
					lastPart := urlParts[len(urlParts)-1]
					products[len(products)-1].Category = strings.Replace(lastPart, "-", " ", -1)
				}
			}
		}
	})

	// Extract specifications from accordion sections
	productCollector.OnHTML(".accordion-item__description", func(e *colly.HTMLElement) {
		if len(products) > 0 {
			specs := strings.TrimSpace(e.Text)
			if specs != "" {
				if products[len(products)-1].Specs == "" {
					products[len(products)-1].Specs = specs
				} else {
					products[len(products)-1].Specs += " | " + specs
				}
			}
		}
	})

	// Extract additional content as specs fallback
	productCollector.OnHTML(".sqs-html-content p", func(e *colly.HTMLElement) {
		if len(products) > 0 && products[len(products)-1].Specs == "" {
			text := strings.TrimSpace(e.Text)
			if text != "" && len(text) < 500 { // Short paragraphs likely specs
				products[len(products)-1].Specs = text
			}
		}
	})

	// Check for custom order info
	productCollector.OnHTML("em:contains('Custom Order')", func(e *colly.HTMLElement) {
		if len(products) > 0 {
			products[len(products)-1].CustomInfo = "Custom Order"
		}
	})

	// Visit all discovered URLs
	for _, url := range knownUrls {
		productCollector.Visit(url)
	}

	file, err := os.Create("products.csv")
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	// Write UTF-8 BOM for Excel compatibility
	_, err = file.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		log.Fatal("Error writing BOM:", err)
	}

	writer := csv.NewWriter(file)

	// Write header
	err = writer.Write([]string{"URL", "Title", "Brand", "Description", "Price", "ImageURL", "Category", "Specs", "CustomInfo"})
	if err != nil {
		log.Fatal("Error writing header:", err)
	}

	// Visit all discovered URLs
	for _, url := range knownUrls {
		productCollector.Visit(url)
	}

	// Write all products to CSV
	for _, product := range products {
		err = writer.Write([]string{
			product.URL,
			product.Title,
			product.Brand,
			product.Description,
			product.Price,
			product.ImageURL,
			product.Category,
			product.Specs,
			product.CustomInfo,
		})
		if err != nil {
			log.Fatal("Error writing product:", err)
		}
	}

	fmt.Println("Collected", len(products), "products")
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Error flushing data to file:", err)
	}

	fmt.Println("Successfully created products.csv")

}
