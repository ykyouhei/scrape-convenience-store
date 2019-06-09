package scraper

import (
	"log"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var areas = map[string]int{
	"北海道":    0,
	"東北":     1,
	"関東":     2,
	"甲信越・北陸": 3,
	"東海":     4,
	"近畿":     5,
	"中国・四国":  6,
	"九州":     7,
}

// SevenElevenItem セブンイレブンの商品データ
type SevenElevenItem struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	ImageURL         string `json:"imageURL"`
	DetailURL        string `json:"detailURL"`
	TaxIncludedPrice int    `json:"taxIncludedPrice"`
	TaxExcludedPrice int    `json:"taxExcludedPrice"`
	LaunchDate       int64  `json:"launchDate"`
	Areas            []int  `json:"areas"`
}

// SevenElevenScraper FamilyMartの商品データリスト
type SevenElevenScraper struct {
}

// DatabasePath FirebaseRTDのパスを返す
func (scraper SevenElevenScraper) DatabasePath() string {
	return "seven_items/thisweek"
}

// Scrape FirebaseRTDに保存するJSONを返す
func (scraper SevenElevenScraper) Scrape() map[string]interface{} {
	log.Println("=============== [START] SevenElevenItems ==============")

	const (
		baseURL  = "http://www.sej.co.jp"
		itemPath = "/i/products/thisweek/"
	)

	items := []SevenElevenItem{}
	doc, _ := htmlquery.LoadURL(baseURL + itemPath)

	areasNodes := htmlquery.Find(doc, "//*[@id='main']//h2")
	itemListNodes := htmlquery.Find(doc, "//*[@id='main']/div[@class='subCategory']//ul[@class='itemList']")

	for index, areaNode := range areasNodes {
		area := areas[htmlquery.InnerText(areaNode.FirstChild)]

		htmlquery.FindEach(itemListNodes[index], "li[@class='item']", func(index int, itemNode *html.Node) {
			summryNode := htmlquery.FindOne(itemNode, "div[@class='summary']")
			itemNameNode := htmlquery.FindOne(summryNode, "div[@class='itemName']")
			itemPriceNode := htmlquery.FindOne(summryNode, "ul[@class='itemPrice']")

			itemName := htmlquery.InnerText(htmlquery.FindOne(itemNameNode, "//a"))
			detailURL := baseURL + htmlquery.SelectAttr(htmlquery.FindOne(itemNameNode, "//a"), "href")
			id := extractItemID(detailURL)
			launchTime := launchTime(htmlquery.InnerText(htmlquery.FindOne(itemPriceNode, "li[@class='launch']")), newJPSeparator())
			taxExcluded, taxIncluded := extractPrices(htmlquery.InnerText(htmlquery.FindOne(itemPriceNode, "li[@class='price']")))
			detailText, imgURL := loadDetailTextAndImageURL(detailURL)

			item := SevenElevenItem{
				id,
				itemName,
				detailText,
				imgURL,
				detailURL,
				taxIncluded,
				taxExcluded,
				launchTime,
				[]int{area}}

			items = update(items, item, area)

			log.Printf("[S]%s: %s\n", item.ID, itemName)
		})
	}

	results := map[string]interface{}{}

	for _, item := range items {
		results[item.ID] = item
	}

	return results
}

// MARK: - Private

func update(items []SevenElevenItem, item SevenElevenItem, area int) []SevenElevenItem {
	newItems := items

	for index, oldItem := range items {
		if oldItem.ID == item.ID {
			item.Areas = append(oldItem.Areas, area)
			newItems[index] = item

			return newItems
		}
	}

	item.Areas = []int{area}

	return append(newItems, item)
}

func loadDetailTextAndImageURL(detailURL string) (detailText, imageURL string) {
	doc, _ := htmlquery.LoadURL(detailURL)
	textNode := htmlquery.FindOne(doc, "//div[@class='text']")
	imgNode := htmlquery.FindOne(doc, "//div[@class='image']")

	a := htmlquery.InnerText(textNode)
	b := htmlquery.SelectAttr(htmlquery.FindOne(imgNode, "//img"), "src")

	return a, "https:" + b
}

func extractItemID(detailURL string) (id string) {
	r := regexp.MustCompile(`/item/([0-9]+).html`)
	id = r.FindAllStringSubmatch(detailURL, -1)[0][1]

	// 先頭2文字は地域コードのため除外
	return id[2:utf8.RuneCountInString(id)]
}

func extractPrices(text string) (taxExcluded, taxIncluded int) {
	r := regexp.MustCompile(`([0-9]+)`)
	prices := r.FindAllStringSubmatch(text, -1)

	taxExcluded, _ = strconv.Atoi(prices[0][0])
	taxIncluded, _ = strconv.Atoi(prices[1][0])

	return
}
