package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/boyppt/notiferhub/aria2rpc"
	"github.com/boyppt/notiferhub/tgbot"
	"github.com/joho/godotenv"
)

func notifyText(path, size string) string {
	sizecnt, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		sizecnt = 0
	}

	return fmt.Sprintf(`*Task Finished*
Torrent: *%s*
Size: *%s*`, path, byteCountSI(sizecnt))
}

func dlURL(path string) string {
	base := os.Getenv("sourceroot")
	return fmt.Sprintf("%s/%s", base, path)
}

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

func tryMax(max int, fun func(string) error, arg string) error {

	var terr error

	for {
		max--
		if max < 0 {
			break
		}

		err := fun(arg)
		if err == nil {
			return nil
		}

		terr = fmt.Errorf("run error > %w", err)
		log.Printf("tryMax got err %v", err)
	}

	return terr
}

func main() {
	homedir, _ := os.UserHomeDir()
	conf := path.Join(homedir, ".ptutils.config")

	err := godotenv.Load(conf)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cldPath := os.Getenv("CLD_PATH")
	cldType := os.Getenv("CLD_TYPE")
	cldSize := os.Getenv("CLD_SIZE")

	switch cldType {
	case "torrent":
		tryMax(3, tgbot.JustNotify, notifyText(cldPath, cldSize))
	case "file":
		if terr := tryMax(10, aria2rpc.JustAddURL, dlURL(cldPath)); terr != nil {
			f, err := os.OpenFile("/tmp/aria2_failing_uris.txt",
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}
			f.Close()
			if _, err := f.WriteString(terr.Error() + "\n"); err != nil {
				log.Println(err)
			}
		}
	default:
	}
}
