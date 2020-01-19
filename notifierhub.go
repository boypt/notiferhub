package notifierhub

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boypt/notiferhub/common"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/hako/durafmt"
	"github.com/spf13/viper"
)

var (
	RedisClient *redis.Client
)

func NewTorrentfromCLD() (*TorrentTask, error) {

	t := &TorrentTask{
		Uuid: uuid.New().String(),
		Path: os.Getenv("CLD_PATH"),
		Type: os.Getenv("CLD_TYPE"),
		Size: os.Getenv("CLD_SIZE"),
		Rest: os.Getenv("CLD_RESTAPI"),
		Hash: os.Getenv("CLD_HASH"),
	}

	if ts := os.Getenv("CLD_STARTTS"); ts != "" {
		if ts, err := strconv.ParseInt(ts, 10, 64); err == nil {
			t.Startts = ts
		}
	}

	return t, nil
}

func (d TorrentTask) DLText() string {

	var dur string
	if d.Startts > 0 {
		dur = durafmt.Parse(time.Since(time.Unix(d.Startts, 0))).LimitFirstN(2).String()
	}
	return fmt.Sprintf(`*%s*
Size: *%s*
Dur: *%v*`, d.Path, d.SizeStr(), dur)
}

func (d TorrentTask) DLURL() []string {
	base := viper.GetStringSlice("sourceroot")
	escaped := url.PathEscape(d.Path)

	var urls []string
	for _, bs := range base {

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

func (d TorrentTask) SizeInt() int64 {
	sizecnt, err := strconv.ParseInt(d.Size, 10, 64)
	if err != nil {
		return -1
	}
	return sizecnt
}

func (d TorrentTask) SizeStr() string {
	return HumaneSize(d.SizeInt())
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

func HumaneSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func init() {

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
