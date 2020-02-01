package ipocalen

import (
	"errors"
	"fmt"
	"strings"

	"net/http"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/PuerkitoBio/goquery"
)

const (
	easturl = "https://data.eastmoney.com/xg/xg/calendar.html"
	mockua  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36"
)

func FetchRootSelection() (*goquery.Selection, error) {
	req, err := http.NewRequest("GET", easturl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", mockua)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	treader := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	doc, err := goquery.NewDocumentFromReader(treader)
	if err != nil {
		return nil, err
	}

	if s := doc.Selection.Find("td.today"); s.Is("td") {
		return s, nil
	}

	return nil, errors.New("today not available")
}

func FindTodayCalendar(sel *goquery.Selection) []string {
	items := sel.Find(".cal_content .cal_item")
	lines := []string{}

	items.Each(func(_ int, si *goquery.Selection) {
		title := si.Children().First().Text()
		switch title {
		case "申 购", "上 市", "缴款日":
			secs := []string{}
			si.Find("ul li").Each(func(_ int, c *goquery.Selection) {
				secs = append(secs, c.Text())
			})
			if len(secs) == 1 && secs[0] == "无" {
				break
			}
			lines = append(lines, fmt.Sprintf("*%s*: %s", title, strings.Join(secs, ",")))
		}
	})
	return lines
}
