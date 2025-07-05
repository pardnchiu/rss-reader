package util

import (
	"net/http"
	"regexp"
	"rss-reader/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Extractor struct {
	client *http.Client
}

func NewExtractor() *Extractor {
	return &Extractor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *Extractor) Get(url string) (*model.NewsContent, error) {
	res, err := e.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	doc.Find("script, style, nav, aside, footer, .advertisement, .ads, .comment").Remove()

	// 檢查 h1 是否存在
	title := doc.Find("h1").First().Text()
	if title == "" {
		// 如果 h1 不存在，則使用 title 標籤
		title = doc.Find("title").Text()
	}

	// 檢查作者信息
	author := doc.Find("[rel='author'], .author, [itemprop='author']").First().Text()

	// 分析主內容
	content := e.getContent(doc)

	return &model.NewsContent{
		Title:     strings.TrimSpace(title),
		Author:    strings.TrimSpace(author),
		Content:   e.clean(content),
		WordCount: e.count(content),
	}, nil
}

func (e *Extractor) getContent(doc *goquery.Document) string {
	// 超過 128 個字符的內容就假設為文章內容
	contentMinLength := 128

	// 檢查 article 標籤
	if content := doc.Find("article").Text(); content != "" && len(content) > contentMinLength {
		return content
	}
	// 檢查 main 標籤
	if content := doc.Find("main").Text(); content != "" && len(content) > contentMinLength {
		return content
	}

	// 檢查全部 div 標籤
	content := ""
	maxLength := 0
	doc.Find("div").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if len(text) > maxLength && len(text) > contentMinLength {
			link := ""
			s.Find("a").Each(func(j int, s *goquery.Selection) {
				link += s.Text()
			})

			percent := float64(len(link)) / float64(len(text))
			if percent < 0.3 {
				maxLength = len(text)
				content = text
			}
		}
	})

	return content
}

func (e *Extractor) clean(str string) string {
	regex := regexp.MustCompile(`\s+`)
	str = regex.ReplaceAllString(str, " ")
	return strings.TrimSpace(str)
}

func (e *Extractor) count(str string) int {
	if str == "" {
		return 0
	}

	cnRegex := regexp.MustCompile(`[\p{Han}]`)
	cnCount := len(cnRegex.FindAllString(str, -1))

	enRegex := regexp.MustCompile(`[a-zA-Z]+`)
	enCount := len(enRegex.FindAllString(str, -1))

	return cnCount + enCount
}
