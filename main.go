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
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"regexp"
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
	AboutBrand  string
	Customize   string
	Specs       string
	CustomInfo  string
	PDFURL      string
}

var products []Product
var pageData map[string]Product

func extractTitle(str string) (string, error) {
	// Extract everything before "by"
	re := regexp.MustCompile(`^(.*?)\s+by\s+`)
	match := re.FindStringSubmatch(str)

	if len(match) > 1 {
		title := strings.TrimSpace(match[1])
		if title != "" {
			fmt.Println("Title:", title)
			return title, nil
		}
	}

	// Fallback: if no "by" found, take everything before " —"
	re2 := regexp.MustCompile(`^(.*?)\s+—`)
	match2 := re2.FindStringSubmatch(str)
	if len(match2) > 1 {
		title := strings.TrimSpace(match2[1])
		if title != "" {
			fmt.Println("Title (fallback):", title)
			return title, nil
		}
	}

	fmt.Println("Title not found")
	return "", errors.New("Title not found in: " + str)
}

func extractBrand(str string) (string, error) {
	// Extract brand between "by" and " —"
	re := regexp.MustCompile(`by\s+(.*?)\s+—`)
	match := re.FindStringSubmatch(str)

	if len(match) > 1 {
		brand := strings.TrimSpace(match[1])
		if brand != "" {
			fmt.Println("Brand:", brand)
			return brand, nil
		}
	}

	// Fallback: extract between "by" and end of string, then take first word
	re2 := regexp.MustCompile(`by\s+(.*)$`)
	match2 := re2.FindStringSubmatch(str)
	if len(match2) > 1 {
		brandPart := strings.TrimSpace(match2[1])
		// Take first word only
		words := strings.Fields(brandPart)
		if len(words) > 0 {
			brand := words[0]
			fmt.Println("Brand (fallback):", brand)
			return brand, nil
		}
	}

	fmt.Println("Brand not found")
	return "", errors.New("Brand not found in: " + str)
}

func extractCategoryFromTitle(title string) string {
	// Define furniture category keywords
	categories := map[string][]string{
		"Chair":   {"chair", "recliner", "accent chair", "wing", "glider"},
		"Sofa":    {"sofa", "sectional", "loveseat", "sofabed"},
		"Table":   {"table", "dining table", "coffee table", "end table"},
		"Bed":     {"bed", "platform", "headboard"},
		"Ottoman": {"ottoman", "footstool"},
		"Bench":   {"bench"},
		"Stool":   {"stool"},
	}

	titleLower := strings.ToLower(title)

	// Check for category keywords in title
	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(titleLower, keyword) {
				return category
			}
		}
	}

	return "Unknown"
}

func getSitemapURLs(website string) ([]string, error) {
	var urls []string

	c := colly.NewCollector(colly.AllowedDomains("www.rkfurnituregallery.ca"))

	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		urls = append(urls, e.Text)
	})

	err := c.Visit(website)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func setupProductCollector() *colly.Collector {
	pageData = make(map[string]Product)

	productCollector := colly.NewCollector(colly.AllowedDomains("www.rkfurnituregallery.ca"))

	// Initialize page data when visiting
	productCollector.OnRequest(func(r *colly.Request) {
		url := r.URL.String()
		pageData[url] = Product{URL: url}
	})

	// Extract title from page title
	productCollector.OnHTML("title", func(e *colly.HTMLElement) {
		fmt.Printf("Processing: %s\n", e.Text)
	})

	// Extract product data from OpenGraph meta tags
	productCollector.OnHTML("meta[property='og:title']", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			product.Title = e.Attr("content")
			pageData[url] = product
		}
	})

	// Extract description
	productCollector.OnHTML("meta[property='og:description']", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			product.Description = e.Attr("content")
			pageData[url] = product
			product.Category = extractCategoryFromTitle(e.Attr("content"))
		}
	})

	// Extract image
	productCollector.OnHTML("meta[property='og:image']", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			product.ImageURL = e.Attr("content")
			pageData[url] = product
		}
	})

	// Extract brand, category, and title from og:title meta tag
	productCollector.OnHTML("meta[property='og:title']", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			title := e.Attr("content")
			product.Title = title

			// Extract title, brand, and category from og:title
			if strings.Contains(title, " by ") {
				parts := strings.Split(title, " by ")
				if len(parts) >= 2 {
					productTitle := strings.TrimSpace(parts[0])
					brandAndSite := strings.TrimSpace(parts[1])

					// Remove site suffix from brand
					brandPart := strings.Split(brandAndSite, " —")
					brand := strings.TrimSpace(brandPart[0])

					product.Brand = brand
				}
			}
			pageData[url] = product
		}
	})

	// Extract price from h1 accent-colored span
	productCollector.OnHTML("h1 .sqsrte-text-color--accent", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			price := strings.TrimSpace(e.Text)
			// Clean price: remove * and ensure proper format
			price = strings.Trim(price, "*")
			if price != "" && price != "$" {
				product.Price = price
				pageData[url] = product
			}
		}
	})

	// Extract PDF
	productCollector.OnHTML(".accordion-item__description a[href$='.pdf']", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			product.PDFURL = e.Attr("href")
			// Convert relative URLs to absolute if needed
			if strings.HasPrefix(product.PDFURL, "/") {
				product.PDFURL = "https://www.rkfurnituregallery.ca" + product.PDFURL
			}
			pageData[url] = product
		}
	})

	// Extract first accordion (About Brand)
	productCollector.OnHTML(".accordion-items-container .accordion-item:nth-child(1) .accordion-item__description", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			aboutBrand := strings.TrimSpace(e.Text)
			if aboutBrand != "" {
				product.AboutBrand = aboutBrand
				pageData[url] = product
			}
		}
	})

	// Extract second accordion (Customize)
	productCollector.OnHTML(".accordion-items-container .accordion-item:nth-child(2) .accordion-item__description", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			customize := strings.TrimSpace(e.Text)
			if customize != "" {
				product.Customize = customize
				pageData[url] = product
			}
		}
	})

	// Extract third accordion (Specifications/Dimensions)
	productCollector.OnHTML(".accordion-items-container .accordion-item:nth-child(3) .accordion-item__description", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			specs := strings.TrimSpace(e.Text)
			if specs != "" {
				product.Specs = specs
				pageData[url] = product
			}
		}
	})

	// Extract additional content as specs fallback
	productCollector.OnHTML(".sqs-html-content p", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists && product.Specs == "" {
			text := strings.TrimSpace(e.Text)
			if text != "" && len(text) < 500 { // Short paragraphs likely specs
				product.Specs = text
				pageData[url] = product
			}
		}
	})

	// Check for custom order info
	productCollector.OnHTML("em:contains('Custom Order')", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if product, exists := pageData[url]; exists {
			product.CustomInfo = "Custom Order"
			pageData[url] = product
		}
	})

	return productCollector
}

func validateAndFilterProducts() {
	for url, product := range pageData {
		// Extract title and brand from the og:title field
		title, titleErr := extractTitle(product.Title)
		brand, brandErr := extractBrand(product.Title)

		if titleErr == nil && brandErr == nil && title != "" && brand != "" {
			product.Title = title
			product.Brand = brand
			products = append(products, product)
			fmt.Printf("Valid product: %s - %s\n", brand, title)
		} else {
			fmt.Printf("Skipping invalid page: %s (Title: %s, Brand: %s)\n", url, product.Title, product.Brand)
		}
	}
}

func writeProductsToCSV() error {
	file, err := os.Create("products.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	// Write UTF-8 BOM for Excel compatibility
	_, err = file.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)

	// Write header
	err = writer.Write([]string{"URL", "Title", "Brand", "Description", "Price", "ImageURL", "Category", "AboutBrand", "Customize", "Specs", "CustomInfo", "PDFURL"})
	if err != nil {
		return err
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
			product.AboutBrand,
			product.Customize,
			product.Specs,
			product.CustomInfo,
			product.PDFURL,
		})
		if err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	return nil
}

func main() {
	const website = "https://www.rkfurnituregallery.ca/sitemap.xml"

	// Get all URLs from sitemap
	knownUrls, err := getSitemapURLs(website)
	if err != nil {
		log.Fatal("Error getting sitemap URLs:", err)
	}

	fmt.Printf("Found %d URLs in sitemap\n", len(knownUrls))

	// Setup product collector
	productCollector := setupProductCollector()

	// Visit all discovered URLs
	for _, url := range knownUrls {
		productCollector.Visit(url)
	}

	// Validate and filter products
	validateAndFilterProducts()

	// Write products to CSV
	err = writeProductsToCSV()
	if err != nil {
		log.Fatal("Error writing CSV:", err)
	}

	fmt.Printf("Successfully created products.csv with %d valid products\n", len(products))
}
