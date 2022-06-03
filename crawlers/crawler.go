package crawlers

import (
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Crawler interface {
	Fetch(*sync.WaitGroup) chan []Product
	GetName() string
	GetContent(int) *goquery.Document
}
