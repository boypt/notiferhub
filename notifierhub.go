//go:generate protoc --go_out=. task.proto
package notifierhub

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boypt/notiferhub/common"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var (
	RedisClient *redis.Client
)

func NewTorrentfromCLD() (*TorrentTask, error) {

	size, err := strconv.ParseInt(os.Getenv("CLD_SIZE"), 10, 64)
	if err != nil {
		log.Println("parse CLD_SIZE error", err)
		size = -1
	}

	t := &TorrentTask{
		Uuid:     uuid.New().String(),
		Path:     os.Getenv("CLD_PATH"),
		Type:     os.Getenv("CLD_TYPE"),
		Size:     size,
		Rest:     os.Getenv("CLD_RESTAPI"),
		Hash:     os.Getenv("CLD_HASH"),
		FinishTS: time.Now().Unix(),
	}

	if ts := os.Getenv("CLD_STARTTS"); ts != "" {
		if ts, err := strconv.ParseInt(ts, 10, 64); err == nil {
			t.StartTS = ts
		}
	}

	return t, nil
}

func (d TorrentTask) DLText() string {

	dur := d.SinceStart()
	var durTxt string
	var avg int64
	if d.StartTS > 0 {
		durTxt = common.KitchenDuration(dur)
		avg = int64(float64(d.Size) / dur.Round(time.Second).Seconds())
	}

	return fmt.Sprintf(`*%s*
Size: *%s*
Dur: *%v*
Avg: *%s/s*`,
		d.Path,
		common.HumaneSize(d.Size),
		durTxt,
		common.HumaneSize(avg))
}

func (d TorrentTask) DLURL() []string {
	base := viper.GetStringSlice("sourceroot")
	escaped := url.PathEscape(d.Path)

	var urls []string
	for _, bs := range base {

		u, err := url.Parse(bs)
		if err != nil {
			log.Println(err)
			continue
		}
		if strings.HasPrefix(u.Hostname(), "drd") {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, u.Hostname())
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("resolve", bs, "->", ips)
			for _, ip := range ips {
				if ip.IP.To4() == nil {
					port := u.Port()
					if port != "" {
						port = ":" + port
					}
					u.Host = "[" + ip.IP.String() + "]" + port
					urls = append(urls, u.String()+escaped)
					continue
				}
			}
		}

		if strings.HasSuffix(bs, "/") {
			urls = append(urls, bs+escaped)
			continue
		}

		if strings.Contains(bs, "?") {
			urls = append(urls, bs+url.QueryEscape(escaped))
			continue
		}

		urls = append(urls, fmt.Sprintf("%s/%s", base, escaped))
	}

	return urls
}

func (d TorrentTask) SinceStart() time.Duration {
	return time.Since(time.Unix(d.StartTS, 0))
}

func (d TorrentTask) failKey() string {
	return fmt.Sprintf("cmddl_fail_%s", d.Uuid)
}

func (d TorrentTask) IsSetFailed() bool {
	ret, err := RedisClient.Exists(d.failKey()).Result()
	common.Must(err)
	return ret == 1
}

func (d TorrentTask) SetFailed() {
	RedisClient.Set(d.failKey(), "set", time.Minute*30)
}

func (d TorrentTask) StopAndRemove() error {
	if d.Type != "torrent" || d.Rest == "" {
		return nil
	}
	if err := CldPOST(d.Rest, "torrent", "stop", d.Hash); err != nil {
		log.Println("CldPOST", err)
	}
	time.Sleep(time.Second)
	if err := CldPOST(d.Rest, "torrent", "delete", d.Hash); err != nil {
		log.Println("CldPOST", err)
	}
	log.Println("[Task StopAndRemoved]", d.Path)
	return nil
}

func CldPOST(host, action string, params ...string) error {
	url := fmt.Sprintf("http://%s/api/%s", host, action)
	ac := strings.Join(params, ":")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(ac)))
	if err != nil {
		log.Println(err)
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	return nil
}

func init() {

	viper.SetDefault("delay_remove", 30)
	viper.SetDefault("redis_addr", "localhost:6379")
	viper.SetDefault("redis_password", "")

	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_addr"),
		Password: viper.GetString("redis_password"),
		DB:       0,
	})

	common.Must2(client.Ping().Result())
	RedisClient = client
}
