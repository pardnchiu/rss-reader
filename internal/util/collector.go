package util

import (
	"encoding/xml"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"rss-reader/internal/database"
	"rss-reader/internal/model"
)

type Collector struct {
	db     *database.SQLite
	client *http.Client
}

func NewCollector(db *database.SQLite) *Collector {
	return &Collector{
		db: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Collector) Add(link string) error {
	if link == "" {
		return nil
	}
	return c.db.InsertFeed(link)
}

func (c *Collector) List() ([]string, error) {
	return c.db.GetFeed()
}

func (c *Collector) Remove(link string) error {
	return c.db.RemoveFeed(link)
}

func (c *Collector) GetNews() ([]model.News, error) {
	feeds, err := c.db.GetFeed()
	if err != nil {
		return nil, err
	}

	var allArticles []model.News
	URLMap := make(map[string]bool)
	now := time.Now()
	threeDay := now.Add(-72 * time.Hour)

	for _, feed := range feeds {
		rss, err := c.fetch(feed)
		if err != nil {
			log.Printf("Failed to get RSS %s: %v", feed, err)
			continue
		}

		source := strings.TrimSpace(rss.Channel.Title)
		if source == "" {
			source = strings.TrimSpace(feed)
		}

		for _, item := range rss.Channel.Item {
			item.Link = strings.TrimSpace(item.Link)
			if URLMap[item.Link] {
				continue
			}
			URLMap[item.Link] = true

			publishedAt := c.parseDate(item.PubDate)
			if publishedAt.Before(threeDay) {
				continue
			}

			article := model.News{
				Title:       c.clean(item.Title),
				Content:     c.clean(item.Description),
				Source:      source,
				URL:         item.Link,
				PublishedAt: publishedAt,
			}
			allArticles = append(allArticles, article)
		}
	}

	sort.Slice(allArticles, func(i, j int) bool {
		return allArticles[i].PublishedAt.After(allArticles[j].PublishedAt)
	})

	return allArticles, nil
}

func (c *Collector) fetch(url string) (*model.RSS, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rss model.RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, err
	}

	return &rss, nil
}

func (c *Collector) parseDate(str string) time.Time {
	if str == "" {
		return time.Now().UTC()
	}

	str = strings.TrimSpace(str)
	arr := []string{
		"Mon,02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
	}

	for _, e := range arr {
		if t, err := time.Parse(e, str); err == nil {
			return t.UTC()
		}
	}

	return time.Now().UTC()
}

func (c *Collector) clean(content string) string {
	regex := regexp.MustCompile(`<[^>]*>`)
	cleaned := regex.ReplaceAllString(content, "")
	return strings.TrimSpace(html.UnescapeString(cleaned))
}
