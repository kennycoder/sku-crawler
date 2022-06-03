package crawlers

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Mediamarkt struct {
}

func (mf Mediamarkt) GetName() string {
	return "Mediamarkt"
}

func (mf Mediamarkt) Fetch(wg *sync.WaitGroup) chan []Product {

	products := make(chan []Product)

	for i := 1; i <= 3; i++ {
		doc := mf.GetContent(i)
		wg.Add(1)

		go func(doc *goquery.Document, page int, ch chan<- []Product) {
			_products := []Product{}

			// Find the product items
			doc.Find("[data-test=\"mms-product-card\"]").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the title and the price
				title := s.Find("[data-test=\"product-title\"]").Text()
				priceStr := s.Find("[data-test=\"mms-unbranded-price\"]").Find("[class^=\"ScreenreaderTextSpan\"]").Text()
				// priceStr = strings.Replace(priceStr, ".", "", -1)

				price, _ := strconv.ParseFloat(priceStr, 32)
				price = price + rand.Float64()*((price+20)-(price-20))
				price = math.Round(price*100) / 100

				product := Product{title, float64(price), mf.GetName()}
				_products = append(_products, product)

				//fmt.Printf("%s: - %f - [%s]\n", product.Name, product.Price, product.Source)
				// Send to BQ
			})

			wg.Done()
			ch <- _products

		}(doc, i, products)

	}

	return products
}

func (mf Mediamarkt) GetContent(page int) *goquery.Document {
	// This is just a dummy URL where I saved the html of the pages. You can put real URL here but it might present captchas as well as other protections
	res, err := http.Get(fmt.Sprintf("https://storage.googleapis.com/gpucrawler-data/mediamarkt/gk-%d.html", page))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}
