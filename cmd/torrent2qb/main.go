package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	noremove = flag.Bool("nr", false, "dont remove torrents")
)

type qbApi struct {
	baseUrl string
	client  *http.Client
}

func (q *qbApi) Login(user, pass string) error {
	api := q.baseUrl + "/api/v2/auth/login"

	if user == "" || pass == "" {
		return fmt.Errorf("User or password is empty")
	}

	resp, err := q.client.PostForm(api, url.Values{
		"username": {user},
		"password": {pass},
	})
	if err != nil {
		return err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	ret := string(body)
	log.Println("Login:", resp.Status, ret)
	if !strings.HasPrefix(ret, "Ok") {
		return fmt.Errorf("Login failed: %s", ret)
	}
	return nil
}

func (q *qbApi) Logout() error {
	api := q.baseUrl + "/api/v2/auth/logout"

	resp, err := q.client.PostForm(api, nil)
	if err != nil {
		return err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	ret := string(body)
	log.Println("Logout:", resp.Status, ret)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Logout error: %s", resp.Status)
	}
	return nil
}

func (q *qbApi) Upload(filename string) error {

	url := q.baseUrl + "/api/v2/torrents/add"
	filetype := "application/x-bittorrent"

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filetype, filepath.Base(file.Name()))
	if err != nil {
		log.Fatal(err)
	}

	io.Copy(part, file)

	if c := ParseCatagory(filename); c != "" {
		log.Printf("Add Catagory:[%s]", c)
		if fw, err := writer.CreateFormField("category"); err == nil {
			fw.Write([]byte(c))
		}
	}

	writer.Close()

	request, err := http.NewRequest("POST", url, body)

	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	response, err := q.client.Do(request)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(response.Status, string(content))
	if response.StatusCode != 200 {
		return fmt.Errorf("Upload failed")
	}
	return nil
}

func ParseCatagory(fn string) string {
	base := filepath.Base(fn)
	if strings.HasPrefix(base, "[GT]") {
		return "PORN"
	}

	if strings.Contains(base, "porn") {
		return "PORN"
	}

	return ""
}

/*
func torrent2cloud(fn string) error {

	buff := &bytes.Buffer{}
	mi, err := metainfo.LoadFromFile(fn)
	if err != nil {
		return err
	}

	if ifo, err := mi.UnmarshalInfo(); err == nil {
		fmt.Println("--> [", ifo.Name, "]", filepath.Base(fn))
	}

	if *del {
		if strings.Contains(mi.Announce, "plab.site") {
			log.Println("remove Annouce: ", mi.Announce)
			mi.Announce = ""
			mi.AnnounceList = metainfo.AnnounceList{}
		}

		if err := mi.Write(buff); err != nil {
			return err
		}
	} else {
		fbyte, err := ioutil.ReadFile(fn)
		if err != nil {
			return err
		}
		buff = bytes.NewBuffer(fbyte)
	}

	return postTorrent(buff)
}
*/

func newQbApi(base string) (*qbApi, error) {

	q := &qbApi{
		baseUrl: strings.TrimSuffix(base, "/"),
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	q.client = &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return q, nil
}

func main() {
	flag.Parse()

	cur, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	_ = godotenv.Load(filepath.Join(cur.HomeDir, ".ptutils.config"))
	_ = godotenv.Load() // for .env

	tors := []string{}
	for _, path := range strings.Split(os.Getenv("CLDTORRENTDIR"), " ") {
		if t, err := filepath.Glob(fmt.Sprintf("%s/*.torrent", path)); err == nil {
			log.Printf("Path %s found %d\n", path, len(t))
			tors = append(tors, t...)
		}
	}

	fmt.Println("===================================")
	q, _ := newQbApi(os.Getenv("QB_BASE_URL"))
	if len(tors) > 0 {
		log.Println("Start login")
		if err := q.Login(os.Getenv("QB_USER"), os.Getenv("QB_PASS")); err != nil {
			log.Fatal(err)
		}
		fmt.Println("===================================")
	}

	for _, torf := range tors {
		for {
			if err := q.Upload(torf); err == nil {
				if !*noremove {
					os.Remove(torf)
					log.Println("removed ", filepath.Base(torf))
				}
				fmt.Println("===================================")
				break
			}
			fmt.Println("err", err)
			time.Sleep(time.Second * 3)
		}
	}

	if len(tors) > 0 {
		log.Println("exit Logout")
		if err := q.Logout(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("===================================")
	}

	if runtime.GOOS == "windows" {
		fmt.Println("===================================")
		fmt.Println("\nPress 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
