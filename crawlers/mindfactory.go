package crawlers

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Mindfactory struct {
}

func (mf Mindfactory) GetName() string {
	return "Mindfactory"
}

func (mf Mindfactory) Fetch(wg *sync.WaitGroup) chan []Product {

	products := make(chan []Product)

	for i := 1; i <= 3; i++ {
		doc := mf.GetContent(i)
		wg.Add(1)

		go func(doc *goquery.Document, page int, ch chan []Product) {
			_products := []Product{}

			// Find the product items
			doc.Find("#bProducts .p").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the title and the price
				title := s.Find(".pname").Text()

				m := regexp.MustCompile("[+-]?([0-9]*[.])?[0-9]+")
				priceStr := m.FindString(s.Find(".pprice").Text())
				priceStr = strings.Replace(priceStr, ".", "", -1)
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

func (mf Mindfactory) GetContent(page int) *goquery.Document {
	res, err := http.Get(fmt.Sprintf("https://crawler-buckets.nikolai.pt/mindfactory/gk-%d.html", page))
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
