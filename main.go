package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/katera/og"
	"github.com/recoilme/graceful"
	"github.com/recoilme/rss2tg/rss2tg"
)

var items map[string]bool

func main() {
	items = make(map[string]bool)

	f, err := os.Open("items")
	defer f.Close()
	if err == nil {
		if err := gob.NewDecoder(f).Decode(&items); err != nil {
			log.Fatal("err restore items", err)
		}
	}
	defer itemsSave(items)

	rssList, err := rss2tg.RssList("rss.txt")
	if err != nil {
		log.Fatal(err)
	}

	words, err := rss2tg.WordsList("words.txt")
	if err != nil {
		log.Fatal(err)
	}

	botkey, err := os.ReadFile("tgbot")
	if err != nil {
		log.Fatal(err)
	}
	botkeys := strings.Split(string(bytes.TrimSpace(botkey)), ":")
	if len(botkeys) != 3 {
		log.Fatal("err n!=3,", botkeys)
	}
	botId := botkeys[0]
	apiKey := botkeys[1]
	channelId := botkeys[2]

	feeds := make(map[string]*rss.Feed)
	for _, rssLink := range rssList {
		feed, err := rss.Fetch(rssLink)
		if err != nil {
			log.Println(rssLink, err)
			continue
		}
		feeds[rssLink] = feed
	}

	for _, feed := range feeds {
		feedCheck(feed, words, botId, apiKey, channelId)
	}

	quit := make(chan os.Signal, 1)
	fallback := func() error {
		itemsSave(items)
		fmt.Println("terminated server")
		return nil
	}
	graceful.Unignore(quit, fallback, graceful.Terminate...)

	go func() {
		for now := range time.Tick(time.Minute * 3) {
			fmt.Println(now)
			itemsSave(items)
			for _, feed := range feeds {
				feed.Update()
				feedCheck(feed, words, botId, apiKey, channelId)
			}
		}
	}()
	select {}
}

func itemsSave(items map[string]bool) {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(items); err != nil {
		panic(err)
	}
	if err := os.WriteFile("items", buf.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}
}

func feedUpdate() string {
	return "1"
}

func feedCheck(feed *rss.Feed, words []string, botId, apiKey, channelId string) {
	if len(feed.Items) > 0 {
		for i := len(feed.Items) - 1; i >= 0; i-- {
			item := feed.Items[i]
			categories := strings.Join(item.Categories, " ")
			// Skip items already known.
			if _, ok := items[item.ID]; ok {
				fmt.Println("continue", item.ID)
				continue
			}
			intersect, err := rss2tg.WordsCheck(item.Title+" "+item.Summary+" "+item.Content+" "+categories, words)
			if err != nil {
				fmt.Println("err ", item.Link, err)
			}
			items[item.ID] = true
			if len(intersect) > 1 {
				//send 2 tg
				//fmt.Println(item.Link, intersect)
				tags := ""
				for _, tag := range intersect {
					tag = strings.Replace(tag, " ", "_", -1)
					tag = strings.Replace(tag, "-", "_", -1)
					tags += "#" + tag + "  "
				}
				_ = tags

				host := item.Link
				u, err := url.Parse(item.Link)
				if err == nil {
					host = u.Host
				}

				descr := ""
				res, err := og.GetOpenGraphFromUrl(item.Link)
				if err == nil {
					descr = res.Description + "\n\n"
				}

				text := fmt.Sprintf("%s\n\n%s%s\n\n<a href=\"%s\">%s</a>",
					item.Title, descr, tags, item.Link, host)
				//fmt.Println(text)
				err = rss2tg.TgTextSend(botId, apiKey, channelId, text)
				if err != nil {
					fmt.Println("err send", item.Link, err)
				}
				//break
			}
		}
	}
}
