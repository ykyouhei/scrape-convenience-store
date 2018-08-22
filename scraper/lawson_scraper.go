package scraper

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// LawsonItem Lawsonの商品データ
type LawsonItem struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	ImageURL         string `json:"imageURL"`
	DetailURL        string `json:"detailURL"`
	TaxIncludedPrice int    `json:"taxIncludedPrice"`
	TaxExcludedPrice int    `json:"taxExcludedPrice"`
	LaunchDate       int64  `json:"launchDate"`
}

// LawsonScraper Lawsonの商品データリスト
type LawsonScraper struct {
}

// DatabasePath FirebaseRTDのパスを返す
func (scraper LawsonScraper) DatabasePath() string {
	return "lawson_items/thisweek"
}

// Scrape サイトをスクレイピングして商品リストを返す
func (scraper LawsonScraper) Scrape() map[string]interface{} {
	log.Println("=============== [START] LawsonScraper ==============")

	const (
		baseURL  = "http://www.lawson.co.jp"
		itemPath = "/recommend/new/index.html"
	)

	baseDoc, _ := htmlquery.LoadURL(baseURL + itemPath)
	meta := htmlquery.SelectAttr(htmlquery.FindOne(baseDoc, "//meta"), "content")
	r := regexp.MustCompile(`URL\=(.+)$`)
	redirectPath := r.FindAllStringSubmatch(meta, -1)[0][1]
	doc, _ := htmlquery.LoadURL(baseURL + redirectPath)

	items := map[string]interface{}{}

	htmlquery.FindEach(doc, "//ul[@class='col-3 heightLineParent']/*", func(index int, itemNode *html.Node) {
		itemName := htmlquery.InnerText(htmlquery.FindOne(itemNode, "//p[@class='ttl']"))
		detailURL := baseURL + htmlquery.SelectAttr(htmlquery.FindOne(itemNode, "a"), "href")
		taxIncluded, _ := strconv.Atoi(strings.Replace(htmlquery.InnerText(htmlquery.FindOne(itemNode, "//p[@class='price']/span[1]")), "円", "", -1))
		launchTime := launchTime(htmlquery.InnerText(htmlquery.FindOne(itemNode, "//p[@class='date']/span")), newDotSeparator())
		itemID := extractID(detailURL)

		detailDoc, _ := htmlquery.LoadURL(detailURL)
		imgURL := htmlquery.SelectAttr(htmlquery.FindOne(detailDoc, "//div[@class='leftBlock']/p[@class='mb05']/img"), "src")
		text := htmlquery.InnerText(htmlquery.FindOne(detailDoc, "//div[@class='rightBlock']/p[@class='text']"))

		item := LawsonItem{
			ID:               itemID,
			Title:            itemName,
			Text:             text,
			ImageURL:         imgURL,
			DetailURL:        detailURL,
			TaxIncludedPrice: taxIncluded,
			TaxExcludedPrice: 0,
			LaunchDate:       launchTime}

		items[item.ID] = item

		log.Printf("[L]%s: %s\n", item.ID, itemName)
	})

	return items
}
