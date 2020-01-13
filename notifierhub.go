package notifierhub

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func byteCountSI(b int64) string {
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

type DLTask struct {
	Path    string
	Size    string
	Type    string
	REST    string
	Hash    string
	StartTs time.Time
}

func NewDLfromCLD() (*DLTask, error) {

	t := &DLTask{
		Path: os.Getenv("CLD_PATH"),
		Type: os.Getenv("CLD_TYPE"),
		Size: os.Getenv("CLD_SIZE"),
		REST: os.Getenv("CLD_RESTAPI"),
		Hash: os.Getenv("CLD_HASH"),
	}

	if ts := os.Getenv("CLD_STARTTS"); ts != "" {
		if ts, err := strconv.ParseInt(ts, 10, 64); err == nil {
			t.StartTs = time.Unix(ts, 0)
		}
	}

	return t, nil
}

func (d DLTask) DLText() string {
	sizecnt, err := strconv.ParseInt(d.Size, 10, 64)
	if err != nil {
		sizecnt = 0
	}

	return fmt.Sprintf(`*%s*
Size: *%s*
Dur: *%v*`, d.Path, byteCountSI(sizecnt), time.Since(d.StartTs))
}

func (d DLTask) DLURL() string {
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

func (d DLTask) SizeInt() int64 {
	sizecnt, err := strconv.ParseInt(d.Size, 10, 64)
	if err != nil {
		return -1
	}
	return sizecnt
}
