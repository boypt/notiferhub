package stock

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func GetSinaStockText(id string) (string, error) {
	if id == "" {
		return "", errors.New("ids empty")
	}
	url := "http://hq.sinajs.cn/list=" + id

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	ret, _ := ioutil.ReadAll(resp.Body)
	db, err := simplifiedchinese.GB18030.NewDecoder().Bytes(ret)
	if err != nil {
		return "", err
	}

	return string(db), nil
}

var (
	ErrMarketClosed = errors.New("market closed")
)

func StockIndexText(stockText string, onlyToday bool) (string, error) {
	head := regexp.MustCompile("var hq_str_.+=\"")
	resp := strings.ReplaceAll(stockText, "\";", "")
	resp = head.ReplaceAllString(resp, "")

	var result []string
	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		q := strings.Split(line, ",")
		if onlyToday && !isToday(q[30]) {
			return "", ErrMarketClosed
		}

		preClose, _ := strconv.ParseFloat(q[2], 32)
		curVal, _ := strconv.ParseFloat(q[3], 32)
		vol, _ := strconv.ParseFloat(q[9], 64)
		volunit := ""
		if vol > 100000000 {
			volunit = "亿"
			vol = vol / 100000000
		}
		diff := curVal - preClose
		diffAct := "升"
		if diff < 0 {
			diffAct = "跌"
		}
		diffPct := (math.Abs(diff) / preClose) * 100

		dis := fmt.Sprintf("*%s(%s%.2f%%)*: `%.2f %.1f` *成交%.0f%s*", q[0], diffAct, diffPct, curVal, diff, vol, volunit)

		result = append(result, dis)
	}

	return strings.Join(result, "\n"), nil
}

func isToday(datestr string) bool {
	d, err := time.Parse("2006-01-02", datestr)
	if err != nil {
		return false
	}
	return time.Now().YearDay() == d.YearDay()
}
