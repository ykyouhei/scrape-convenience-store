package scraper

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// FamilyMartItem FamilyMartの商品データ
type FamilyMartItem struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	ImageURL         string `json:"imageURL"`
	DetailURL        string `json:"detailURL"`
	TaxIncludedPrice int    `json:"taxIncludedPrice"`
	TaxExcludedPrice int    `json:"taxExcludedPrice"`
	LaunchDate       int64  `json:"launchDate"`
	Category         string `json:"category"`
}

// FamilyMartScraper FamilyMartの商品データリスト
type FamilyMartScraper struct {
}

// DatabasePath FirebaseRTDのパスを返す
func (scraper FamilyMartScraper) DatabasePath() string {
	return "familymart_items/thisweek"
}

// Scrape サイトをスクレイピングして商品リストを返す
func (scraper FamilyMartScraper) Scrape() map[string]interface{} {
	log.Println("=============== [START] FamilyMartScraper ==============")

	const (
		baseURL  = "http://www.family.co.jp"
		itemPath = "/goods/newgoods.html"
	)

	doc, _ := htmlquery.LoadURL(baseURL + itemPath)
	items := map[string]interface{}{}

	htmlquery.FindEach(doc, "//div[@class='ly-mod-layout-clm']", func(index int, itemNode *html.Node) {
		imgURL := baseURL + htmlquery.SelectAttr(htmlquery.FindOne(itemNode, "//img[@class='ly-hovr']"), "src")
		detailURL := baseURL + htmlquery.SelectAttr(htmlquery.FindOne(itemNode, "//a[@class='ly-mod-infoset4-link']"), "href")
		itemID := extractID(detailURL)
		itemName := strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(itemNode, "//h3[@class='ly-mod-infoset4-ttl']")))
		category := strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(itemNode, "//p[@class='ly-mod-infoset4-cate']")))
		taxExcluded, taxIncluded := extractPrice(htmlquery.InnerText(htmlquery.FindOne(itemNode, "//p[@class='ly-mod-infoset4-txt']")))
		text, launchTime := loadDetail(detailURL)

		item := FamilyMartItem{
			ID:               itemID,
			Title:            itemName,
			Text:             text,
			ImageURL:         imgURL,
			DetailURL:        detailURL,
			TaxIncludedPrice: taxIncluded,
			TaxExcludedPrice: taxExcluded,
			LaunchDate:       launchTime,
			Category:         category}

		items[item.ID] = item

		log.Printf("[F]%s: %s\n", item.ID, itemName)
	})

	return items
}

// MARK: - Private

func loadDetail(detailURL string) (text string, launch int64) {
	doc, _ := htmlquery.LoadURL(detailURL)

	text = strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(doc, "//p[@class='ly-goods-lead']")))
	launch = launchTime(htmlquery.InnerText(htmlquery.FindOne(doc, "//ul[@class='ly-goods-spec']/li")), newJPSeparator())

	return
}

// 税別・税抜き価格を取り出す
func extractPrice(text string) (taxExcluded, taxIncluded int) {
	r := regexp.MustCompile(`([0-9]+)円`)
	prices := r.FindAllStringSubmatch(text, -1)

	taxExcluded, _ = strconv.Atoi(prices[0][1])
	taxIncluded, _ = strconv.Atoi(prices[1][1])

	return
}
