package notifierhub

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/hako/durafmt"
)

var (
	RedisClient *redis.Client
)

func NewTorrentfromCLD() (*TorrentTask, error) {

	t := &TorrentTask{
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
	dur := time.Since(time.Unix(d.Startts, 0))
	return fmt.Sprintf(`*%s*
Size: *%s*
Dur: *%v*`, d.Path, d.SizeStr(), durafmt.Parse(dur))
}

func (d TorrentTask) DLURL() string {
	base := os.Getenv("sourceroot")
	escaped := url.PathEscape(d.Path)

	if strings.HasSuffix(base, "/") {
		return base + escaped
	}

	if strings.Contains(base, "?") {
		return base + url.QueryEscape(escaped)
	}

	return fmt.Sprintf("%s/%s", base, escaped)
}

func (d TorrentTask) SizeInt() int64 {
	sizecnt, err := strconv.ParseInt(d.Size, 10, 64)
	if err != nil {
		return -1
	}
	return sizecnt
}

func (d TorrentTask) SizeStr() string {
	b := d.SizeInt()
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func init() {
	redisconn := os.Getenv("REDISCONN")
	if redisconn == "" {
		redisconn = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisconn,
		DB:   0,
	})

	if _, err := client.Ping().Result(); err != nil {
		fmt.Println(err)
		return
	}

	RedisClient = client
}
