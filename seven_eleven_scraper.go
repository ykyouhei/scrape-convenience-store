package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/antchfx/htmlquery"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/html"
)

const (
	baseURL  = "http://www.sej.co.jp"
	itemPath = "/i/products/thisweek/"
)

// SevenElevenItem セブンイレブンの商品データ
type SevenElevenItem struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Text             string   `json:"text"`
	ImageURL         string   `json:"imageURL"`
	DetailURL        string   `json:"detailURL"`
	TaxIncludedPrice int      `json:"taxIncludedPrice"`
	TaxExcludedPrice int      `json:"taxExcludedPrice"`
	LaunchDate       int64    `json:"launchDate"`
	Areas            []string `json:"areas"`
}

// ScrapeToSevenEleven セブンイレブンの商品ページをスクレイピングする
func ScrapeToSevenEleven() {

	items := []SevenElevenItem{}
	doc, _ := htmlquery.LoadURL(baseURL + itemPath)

	areasNodes := htmlquery.Find(doc, "//*[@id='main']//h2")
	itemListNodes := htmlquery.Find(doc, "//*[@id='main']/div[@class='subCategory']//ul[@class='itemList']")

	for index, areaNode := range areasNodes {
		area := htmlquery.InnerText(areaNode.FirstChild)
		fmt.Printf("=============== %s ==============\n", area)

		htmlquery.FindEach(itemListNodes[index], "li[@class='item']", func(index int, itemNode *html.Node) {
			imgNode := htmlquery.FindOne(itemNode, "div[@class='image']")
			summryNode := htmlquery.FindOne(itemNode, "div[@class='summary']")
			itemNameNode := htmlquery.FindOne(summryNode, "div[@class='itemName']")
			itemPriceNode := htmlquery.FindOne(summryNode, "ul[@class='itemPrice']")

			imgURL := htmlquery.SelectAttr(htmlquery.FindOne(imgNode, "//img"), "src")
			itemName := htmlquery.InnerText(htmlquery.FindOne(itemNameNode, "//a"))
			detailURL := baseURL + htmlquery.SelectAttr(htmlquery.FindOne(itemNameNode, "//a"), "href")
			id := extractItemID(detailURL)
			launchTime := launchTime(htmlquery.InnerText(htmlquery.FindOne(itemPriceNode, "li[@class='launch']")))
			taxExcluded, taxIncluded := extractPrices(htmlquery.InnerText(htmlquery.FindOne(itemPriceNode, "li[@class='price']")))
			detailText := loadDetailText(detailURL)

			item := SevenElevenItem{
				id,
				itemName,
				detailText,
				imgURL,
				detailURL,
				taxIncluded,
				taxExcluded,
				launchTime,
				[]string{area}}

			items = update(items, item, area)
		})
	}

}

// MARK: - Private

func update(items []SevenElevenItem, item SevenElevenItem, area string) []SevenElevenItem {
	newItems := items

	for index, oldItem := range items {
		if oldItem.ID == item.ID {
			item.Areas = append(oldItem.Areas, area)
			newItems[index] = item
			fmt.Print("========= updated: \n")
			spew.Dump(item)

			return newItems
		}
	}

	item.Areas = []string{area}
	fmt.Print("========= new: \n")
	spew.Dump(item)

	return append(newItems, item)
}

func loadDetailText(detailURL string) string {
	doc, _ := htmlquery.LoadURL(detailURL)
	textNode := htmlquery.FindOne(doc, "//div[@class='text']")
	return htmlquery.InnerText(textNode)
}

func launchTime(text string) (unix int64) {
	r := regexp.MustCompile(`([0-9]+)年([0-9]+)月([0-9]+)日`)
	days := r.FindAllStringSubmatch(text, -1)[0]

	year, _ := strconv.Atoi(days[1])
	month, _ := strconv.Atoi(days[2])
	day, _ := strconv.Atoi(days[3])
	jst, _ := time.LoadLocation("Asia/Tokyo")

	unix = time.Date(year, time.Month(month), day, 0, 0, 0, 0, jst).Unix()
	return
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
