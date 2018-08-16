package scraper

import (
	"regexp"
	"strconv"
	"time"
)

// Scraper コンビニ商品リストのスクレイピングを行うインタフェース
type Scraper interface {
	DatabasePath() string
	Scrape() map[string]interface{}
}

type dateSeparator struct {
	year  string
	month string
	day   string
}

func newJPSeparator() dateSeparator {
	return dateSeparator{"年", "月", "日"}
}

func newDotSeparator() dateSeparator {
	return dateSeparator{".", ".", ""}
}

func launchTime(text string, s dateSeparator) (unix int64) {
	r := regexp.MustCompile(`([0-9]+)` + s.year + `([0-9]+)` + s.month + `([0-9]+)` + s.day)
	days := r.FindAllStringSubmatch(text, -1)[0]

	year, _ := strconv.Atoi(days[1])
	month, _ := strconv.Atoi(days[2])
	day, _ := strconv.Atoi(days[3])
	jst, _ := time.LoadLocation("Asia/Tokyo")

	unix = time.Date(year, time.Month(month), day, 0, 0, 0, 0, jst).Unix()
	return
}

// リンクから商品IDを抜き出す
func extractID(link string) string {
	r := regexp.MustCompile(`([0-9|_]+).html`)
	return r.FindAllStringSubmatch(link, -1)[0][1]
}
