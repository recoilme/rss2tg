package rss2tg

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/SlyMarbo/rss"
)

func init() {
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = time.Duration(time.Second * 10)
}

// RssList return array links on rss
// or error if link nor parcelable
func RssList(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(txt, "http") {
			txt = "http://" + txt
		}
		u, err := url.ParseRequestURI(txt)
		if err != nil {
			return lines, err
		}
		lines = append(lines, u.String())
	}
	return lines, scanner.Err()
}

// WordsList return lowercase words
func WordsList(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if txt == "" {
			continue
		}
		lines = append(lines, txt)
	}
	return lines, scanner.Err()
}

func FeedItems(u string, words []string) error {
	feed, err := rss.Fetch(u)
	if err != nil {
		return err
	}
	for _, i := range feed.Items {
		WordsCheck(i.Title+" "+i.Summary+" "+i.Content, words)
	}
	return nil
}

func WordsCheck(txt string, words []string) (intersect []string, err error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(txt))
	if err != nil {
		return
	}
	txt = strings.ToLower(doc.Text())
	fields := strings.Fields(txt)
	txt = strings.Join(fields, " ")
	for _, w := range words {
		if strings.Contains(w, " ") {
			// Ad Tech
			if strings.Contains(txt, w) {
				intersect = append(intersect, w)
			}
		} else {
			// adtech
			for _, f := range fields {
				if f == w {
					intersect = append(intersect, w)
					break
				}
			}
		}
	}
	return
}

func TgTextSend(botId, apiKey, chatId, text string) error {
	link := "https://api.telegram.org/bot{botId}:{apiKey}/sendMessage?chat_id={chatId}&text={text}&disable_web_page_preview=1&parse_mode=HTML"
	link = strings.Replace(link, "{botId}", botId, -1)
	link = strings.Replace(link, "{apiKey}", apiKey, -1)
	link = strings.Replace(link, "{chatId}", chatId, -1)
	link = strings.Replace(link, "{text}", url.QueryEscape(text), -1)

	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		fmt.Println("err ", string(b), err, resp.StatusCode)
		if resp.StatusCode == 429 {
			//too many requests
			time.Sleep(1 * time.Minute)
		}
		return err
	}
	_, err = io.Copy(ioutil.Discard, resp.Body)
	time.Sleep(2 * time.Second)
	return err
}
