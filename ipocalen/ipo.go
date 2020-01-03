package ipocalen

import (
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
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(treader)
	if err != nil {
		return nil, err
	}

	return doc.Selection.Find("td.today"), nil
}

func FindTodayCalendar(sel *goquery.Selection) []string {
	items := sel.Find(".cal_content .cal_item")
	sg := items.First()
	title := sg.Children().First()
	ul := sg.Find("ul li")
	lines := []string{title.Text()}
	ul.Each(func(ix int, s *goquery.Selection) {
		lines = append(lines, s.Text())
	})
	return lines
}
