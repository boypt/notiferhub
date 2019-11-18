package stock

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
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
	ErrNoResult     = errors.New("No result")
)

type Price struct {
	Name       string
	PreClose   float64
	CurValue   float64
	Diff       float64
	DiffSign   string
	DiffDesc   string
	DiffPct    float64
	Volume     float64
	VolumeUnit string
}

func (p *Price) Display() string {
	dis := fmt.Sprintf("*%s(%s%.2f%%)*: `%.2f %s%.1f` *成交%.0f%s*",
		p.Name, p.DiffDesc, p.DiffPct, p.CurValue, p.DiffSign, p.Diff, p.Volume, p.VolumeUnit)

	return dis
}

func parseASidx(val string, onlyToday, debug bool) (*Price, error) {
	q := strings.Split(val, ",")
	if onlyToday && !isToday("2006-01-02", q[30]) {
		return nil, ErrMarketClosed
	}

	if !debug && !timeIsATradeTime() {
		return nil, ErrMarketClosed
	}

	p := &Price{}

	p.Name = q[0]

	if preClose, err := strconv.ParseFloat(q[2], 64); err == nil {
		p.PreClose = preClose
	}

	if curVal, err := strconv.ParseFloat(q[3], 64); err == nil {
		p.CurValue = curVal
	}

	if vol, err := strconv.ParseFloat(q[9], 64); err == nil {
		p.Volume = vol
	}

	if p.Volume > 100000000 {
		p.VolumeUnit = "亿"
		p.Volume = p.Volume / 100000000
	}

	p.Diff = p.CurValue - p.PreClose
	p.DiffDesc = "升"
	p.DiffSign = "+"
	if p.Diff < 0 {
		p.DiffDesc = "跌"
		p.DiffSign = ""
	}
	p.DiffPct = (math.Abs(p.Diff) / p.PreClose) * 100

	return p, nil
}

func parseHKidx(val string, onlyToday, debug bool) (*Price, error) {
	q := strings.Split(val, ",")
	if onlyToday && !isToday("2006/01/02", q[17]) {
		return nil, ErrMarketClosed
	}

	if !debug && !timeIsHKTradeTime() {
		return nil, ErrMarketClosed
	}

	p := &Price{}
	p.Name = q[1]
	if preClose, err := strconv.ParseFloat(q[3], 64); err == nil {
		p.PreClose = preClose
	}

	if curVal, err := strconv.ParseFloat(q[6], 64); err == nil {
		p.CurValue = curVal
	}

	if vol, err := strconv.ParseFloat(q[11], 64); err == nil {
		p.Volume = vol
	}

	if p.Volume > 100000 {
		// 单位： thousand
		p.VolumeUnit = "亿"
		p.Volume = p.Volume / 100000
	}

	if diff, err := strconv.ParseFloat(q[7], 64); err == nil {
		p.Diff = diff
	}

	if diffPct, err := strconv.ParseFloat(q[8], 64); err == nil {
		p.DiffPct = diffPct
	}

	p.DiffDesc = "升"
	p.DiffSign = "+"
	if p.Diff < 0 {
		p.DiffDesc = "跌"
		p.DiffSign = ""
	}

	return p, nil
}

func StockIndexText(stockText string, tradeonly, debug bool) (string, error) {
	resp := strings.ReplaceAll(stockText, "var hq_str_", "")
	resp = strings.ReplaceAll(resp, "\";", "")

	var prices []*Price
	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		kv := strings.Split(line, "=\"")
		k := kv[0]

		if strings.HasPrefix(k, "hk") || strings.HasPrefix(k, "rt_hk") {
			if p, err := parseHKidx(kv[1], tradeonly, debug); err == nil {
				prices = append(prices, p)
			}
		} else if strings.HasPrefix(k, "sh") || strings.HasPrefix(k, "sz") {
			if p, err := parseASidx(kv[1], tradeonly, debug); err == nil {
				prices = append(prices, p)
			}
		}
	}

	if len(prices) > 0 {
		var r []string
		for _, p := range prices {
			r = append(r, p.Display())
		}
		return strings.Join(r, "\n"), nil
	}

	return "", ErrNoResult
}

func isToday(formstr, datestr string) bool {
	d, err := time.Parse(formstr, datestr)
	if err != nil {
		return false
	}
	return time.Now().YearDay() == d.YearDay()
}

func timeIsATradeTime() bool {
	now := time.Now()
	h, m, _ := now.Clock()
	if h >= 9 && h <= 15 {
		if h == 9 && m < 20 {
			return false
		}
		if h == 15 && m > 1 {
			return false
		}
		return true
	}
	return false
}

func timeIsHKTradeTime() bool {
	now := time.Now()
	h, m, _ := now.Clock()
	if h >= 9 && h <= 16 {
		if h == 9 && m < 30 {
			return false
		}
		if h == 16 && m > 10 {
			return false
		}
		return true
	}
	return false
}
