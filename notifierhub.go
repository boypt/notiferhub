package notifierhub

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
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
	Path string
	Size string
	Type string
	REST string
	Hash string
}

func NewDLfromCLD() (*DLTask, error) {

	t := &DLTask{
		Path: os.Getenv("CLD_PATH"),
		Type: os.Getenv("CLD_TYPE"),
		Size: os.Getenv("CLD_SIZE"),
		REST: os.Getenv("CLD_RESTAPI"),
		Hash: os.Getenv("CLD_HASH"),
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
Time: *%s*`, d.Path, byteCountSI(sizecnt), time.Now().Format(time.Stamp))
}

func (d DLTask) DLURL() string {
	base := os.Getenv("sourceroot")
	return fmt.Sprintf("%s/%s", base, url.PathEscape(d.Path))
}

func (d DLTask) SizeInt() int64 {
	sizecnt, err := strconv.ParseInt(d.Size, 10, 64)
	if err != nil {
		return -1
	}
	return sizecnt
}
