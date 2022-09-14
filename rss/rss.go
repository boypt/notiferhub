package rss

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	notifierhub "notiferhub"

	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

const (
	redisRssLastIDs = "cld_rss_lastids"
	cldRestHOST     = "127.0.0.1:1301"
)

var (
	magnetExp  = regexp.MustCompile(`magnet:[^< ]+`)
	httpclient = &http.Client{
		Timeout: 60 * time.Second,
	}
)

func FindFromRSS() {

	fp := gofeed.NewParser()
	fp.Client = httpclient

	var rssResults []*gofeed.Item
	keywords := viper.GetStringSlice("autokeywords")
	// log.Println("keywords", keywords)
	for _, rss := range viper.GetStringSlice("autorss") {

		rss = strings.TrimSpace(rss)
		if !strings.HasPrefix(rss, "http://") && !strings.HasPrefix(rss, "https://") {
			continue
		}
		feed, err := fp.ParseURL(rss)
		if err != nil {
			log.Println("parse feed err", err)
			continue
		}

		// log.Printf("retrived feed %s", feed.Title)
		oldmark, err := notifierhub.RedisClient.HGet(redisRssLastIDs, rss).Result()
		if oldmark != "" && err == nil {

			if len(feed.Items) > 0 && feed.Items[0].GUID != oldmark {
				notifierhub.RedisClient.HSet(redisRssLastIDs, rss, feed.Items[0].GUID)
			}

			var lastIdx int
			for i, item := range feed.Items {
				if item.GUID == oldmark {
					lastIdx = i
					break
				}
			}
			if lastIdx > 0 {
				log.Printf("feed updated with %d new items", lastIdx)
				rssResults = feed.Items[:lastIdx]
			}
		} else if len(feed.Items) > 0 {
			notifierhub.RedisClient.HSet(redisRssLastIDs, rss, feed.Items[0].GUID)
			rssResults = feed.Items
		}

		for _, item := range rssResults {
			// skip old items
			if time.Since(*item.PublishedParsed) > time.Hour*24 {
				continue
			}

			for _, keyw := range keywords {
				if strings.Contains(item.Title, keyw) {
					ltype, link := findLink(item)
					switch ltype {
					case "magnet":
						if err := notifierhub.CldPOST(cldRestHOST, "magnet", link); err != nil {
							log.Println("CldPOST", err)
						}
					case "torrent":
						if err := notifierhub.CldPOST(cldRestHOST, "url", link); err != nil {
							log.Println("CldPOST", err)
						}
					default:
					}
				}
			}
		}
	}
}

func findLink(i *gofeed.Item) (ltype string, link string) {
	ltype = "magnet"
	if etor, ok := i.Extensions["torrent"]; ok {
		if e, ok := etor["magnetURI"]; ok && len(e) > 0 {
			link = e[0].Value
			return
		}
		if e, ok := etor["infoHash"]; ok && len(e) > 0 {
			link = "magnet:?xt=urn:btih:" + e[0].Value
			return
		}
	} else {
		// some sites put it under enclosures
		for _, e := range i.Enclosures {
			if e.Type == "application/x-bittorrent" {
				ltype = "torrent"
				link = e.URL
				return
			}
			if strings.HasPrefix(e.URL, "magnet:") {
				link = e.URL
			}
		}

		// find magnet in description
		if s := magnetExp.FindString(i.Description); s != "" {
			link = s
			return
		}

	}

	ltype = "notfound"
	return
}
