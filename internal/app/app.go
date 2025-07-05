package app

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"rss-reader/internal/database"
	"rss-reader/internal/model"
	"rss-reader/internal/util"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app              *tview.Application
	collector        *util.Collector
	extractor        *util.Extractor
	database         *database.SQLite
	list             *tview.List
	llmView          *tview.TextView
	preview          *tview.TextView
	input            *tview.InputField
	status           *tview.TextView
	articles         []model.News
	filteredArticles []model.News
	ticker           *time.Ticker
	stopChan         chan bool
	autoRefresh      bool
}

func New() *App {
	db, err := database.NewSQLite()
	if err != nil {
		log.Fatalf("Failed to init SQLite: %v", err)
	}

	app := &App{
		app:         tview.NewApplication(),
		collector:   util.NewCollector(db),
		extractor:   util.NewExtractor(),
		database:    db,
		ticker:      time.NewTicker(5 * time.Minute),
		stopChan:    make(chan bool),
		autoRefresh: true,
	}
	app.frame()
	app.refresh()
	return app
}

func (a *App) frame() {
	a.status = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	a.list = tview.NewList().
		ShowSecondaryText(true)
	a.list.SetBorder(true).
		SetTitle("News List").
		SetTitleAlign(tview.AlignLeft)
	a.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(a.filteredArticles) && a.app.GetFocus() == a.list {
			go a.showPreview(a.filteredArticles[index])
		}
	})

	// 後續要添加 llm 分析
	a.llmView = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)
	a.llmView.SetBorder(true).
		SetTitle("Summary").
		SetTitleAlign(tview.AlignLeft)

	leftView := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.list, 0, 1, true)

	a.preview = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)
	a.preview.SetBorder(true).
		SetTitle("Preview").
		SetTitleAlign(tview.AlignLeft)

	a.input = tview.NewInputField().
		SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetFieldTextColor(tview.Styles.PrimaryTextColor)
	a.input.SetBorder(true).
		SetTitle("Command").
		SetTitleAlign(tview.AlignLeft)

	rightView := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.preview, 0, 1, true).
		AddItem(a.input, 3, 0, false)

	mainFlex := tview.NewFlex().
		AddItem(leftView, 0, 1, true).
		AddItem(rightView, 0, 1, false)

	rootFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.status, 2, 0, false).
		AddItem(mainFlex, 0, 1, true)

	a.app.SetRoot(rootFlex, true)
	a.app.SetFocus(a.list)
	a.listener()
}

func (a *App) refresh() {
	go func() {
		for {
			select {
			case <-a.ticker.C:
				if a.autoRefresh {
					a.app.QueueUpdateDraw(func() {
						a.updateStatus("Refreshing...")
					})
					a.getList(true)
				}
			case <-a.stopChan:
				return
			}
		}
	}()
}

func (a *App) listener() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if a.app.GetFocus() == a.list {
				a.app.SetFocus(a.preview)
			} else if a.app.GetFocus() == a.preview {
				a.app.SetFocus(a.input)
			} else if a.app.GetFocus() == a.input {
				// a.app.SetFocus(a.llmView)
				a.app.SetFocus(a.list)
			}
			return nil
		case tcell.KeyCtrlR:
			a.getList(true)
			return nil
		case tcell.KeyCtrlO:
			index := a.list.GetCurrentItem()
			if index >= 0 && index < len(a.filteredArticles) {
				go a.openBrowser(a.filteredArticles[index].URL)
			}
			return nil
		}
		return event
	})

	a.input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			command := a.input.GetText()
			a.input.SetText("")
			a.command(command)
		}
	})
}

func (a *App) command(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	parts := strings.Fields(command)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "add":
		if len(parts) < 7 {
			a.showCommand("add [URL]")
			return
		}
		url := parts[1]
		a.collector.Add(url)
		a.showCommand(fmt.Sprintf("Add RSS: %s", url))
		a.showFeedList()

	case "rm", "remove":
		if len(parts) < 2 {
			a.showCommand("rm [URL]")
			return
		}
		url := parts[1]
		a.collector.Remove(url)
		a.showCommand(fmt.Sprintf("Remove RSS: %s", url))
		a.showFeedList()

	case "list":
		a.showFeedList()

	default:
		a.showCommand(fmt.Sprintf("未知指令: %s", cmd))
	}
}

func (a *App) showFeedList() {
	feeds, err := a.collector.List()
	if err != nil {
		a.showCommand(fmt.Sprintf("Failed to list RSS feed: %v", err))
		return
	}

	if len(feeds) == 0 {
		a.showCommand("No RSS feeds available.")
		return
	}

	result := "RSS feed list:\n"
	for i, feed := range feeds {
		result += fmt.Sprintf("%d. %s\n", i+1, feed)
	}
	a.showCommand(result)
}

func (a *App) showCommand(message string) {
	a.preview.SetText(message).ScrollToBeginning()
}

func (a *App) stopRefresh() {
	if a.ticker != nil {
		a.ticker.Stop()
	}
	select {
	case a.stopChan <- true:
	default:
	}
}

func (a *App) getList(isRefresh bool) {
	if isRefresh {
		a.updateStatus("Checking...")
	} else {
		a.updateStatus("Loading...")
	}

	go func() {
		var articles = a.articles
		storedArticles, err := a.database.Get(72)
		storedMap := make(map[string]model.News)

		if len(a.articles) < 1 {
			if err == nil && len(storedArticles) > 0 {
				for _, stored := range storedArticles {
					storedMap[stored.URL] = stored
				}

				a.app.QueueUpdateDraw(func() {
					a.articles = storedArticles
					a.filteredArticles = a.articles
					a.updateList()
					a.updateStatus(fmt.Sprintf("Get %d news from Database", len(a.articles)))
				})
			}
		}

		articles, err = a.collector.GetNews()
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.updateStatus(fmt.Sprintf("Failed to get news list: %v", err))
			})
			return
		}

		mergedArticles := make([]model.News, 0)
		for _, article := range articles {
			if stored, exists := storedMap[article.URL]; exists {
				article.PublishedAt = stored.PublishedAt
			}
			mergedArticles = append(mergedArticles, article)
		}

		newCount := a.count(articles)
		go a.loadContent(articles)

		a.app.QueueUpdateDraw(func() {
			a.articles = articles
			a.filteredArticles = articles
			a.updateList()
			if newCount > 0 {
				a.updateStatus(fmt.Sprintf("Found %d news, getting full content...", newCount))
			} else {
				a.updateStatus("All news are up to date.")
			}
		})
	}()
}

func (a *App) count(news []model.News) int {
	newCount := 0
	for _, article := range news {
		_, err := a.database.GetFromURL(article.URL)
		if err != nil {
			newCount++
		}
	}
	return newCount
}

func (a *App) loadContent(news []model.News) {
	for i, article := range news {
		stored, err := a.database.GetFromURL(article.URL)
		if err == nil && stored.FullContent != nil {
			continue
		}

		extracted, err := a.extractor.Get(article.URL)
		if err != nil {
			log.Printf("Failed to get content %s: %v", article.URL, err)

			if err := a.database.Insert(article, nil); err != nil {
				log.Printf("Failed to store news %s: %v", article.URL, err)
			}
			continue
		}

		if err := a.database.Insert(article, extracted); err != nil {
			log.Printf("Failed to store news %s: %v", article.URL, err)
		}

		progress := float64(i+1) / float64(len(news)) * 100
		a.app.QueueUpdateDraw(func() {
			a.updateStatus(fmt.Sprintf("Progress: %.1f%% (%d/%d)", progress, i+1, len(news)))
		})

		time.Sleep(500 * time.Millisecond)
	}

	a.app.QueueUpdateDraw(func() {
		a.updateStatus("All news are up to date")
	})
}

func (a *App) updateList() {
	a.list.Clear()

	for _, e := range a.filteredArticles {
		title := e.Title
		timeStr := e.PublishedAt.Local().Format("01/02 15:04")
		date := fmt.Sprintf("%s | %s", timeStr, e.Source)

		a.list.AddItem(title, date, 0, nil)
	}
}

func (a *App) showPreview(news model.News) {
	a.app.QueueUpdateDraw(func() {
		a.preview.SetText("[yellow]Loading...[white]")
	})

	stored, err := a.database.GetFromURL(news.URL)
	if err == nil && stored.FullContent != nil {
		a.app.QueueUpdateDraw(func() {
			author := ""
			if stored.Author != nil {
				author = *stored.Author
			}

			content := ""
			if stored.FullContent != nil {
				content = *stored.FullContent
			}

			wordCount := 0
			if stored.WordCount != nil {
				wordCount = *stored.WordCount
			}

			extracted := &model.NewsContent{
				Title:     stored.Title,
				Author:    author,
				Content:   content,
				WordCount: wordCount,
			}
			a.showFull(news, extracted)
		})
		return
	}

	extracted, err := a.extractor.Get(news.URL)
	if err != nil {
		a.app.QueueUpdateDraw(func() {
			a.showBasicPreview(news)
		})
		return
	}

	go a.database.Insert(news, extracted)

	a.app.QueueUpdateDraw(func() {
		a.showFull(news, extracted)
	})
}

func (a *App) showFull(news model.News, extracted *model.NewsContent) {
	content := fmt.Sprintf("[yellow::b]%s[white::-]\n\n", news.Title)

	if extracted.Author != "" {
		content += fmt.Sprintf("[lightblue]Author:[white] %s\n", extracted.Author)
	}

	content += fmt.Sprintf("[lightblue]Source:[white] %s\n", news.Source)
	content += fmt.Sprintf("[lightblue]Publish:[white] %s\n", news.PublishedAt.Local().Format("2006-01-02 15:04"))

	if extracted.WordCount > 0 {
		content += fmt.Sprintf("[lightblue]Count:[white] %d\n", extracted.WordCount)
	}

	content += fmt.Sprintf("[lightblue]Link:[white] %s\n\n", news.URL)
	content += fmt.Sprintf("[lime]Content:[white]\n%s", a.wrapText(extracted.Content, 80))

	a.preview.SetText(strings.TrimSpace(content)).ScrollToBeginning()
}

func (a *App) showBasicPreview(news model.News) {
	content := fmt.Sprintf("[yellow::b]%s[white::-]\n\n", news.Title)
	content += fmt.Sprintf("[lightblue]Source:[white] %s\n", news.Source)
	content += fmt.Sprintf("[lightblue]Publish:[white] %s\n", news.PublishedAt.Local().Format("2006-01-02 15:04"))
	content += fmt.Sprintf("[lightblue]Link:[white] %s\n\n", news.URL)
	content += fmt.Sprintf("[lime]Summary:[white]\n%s", a.wrapText(strings.TrimSpace(news.Content), 80))

	a.preview.SetText(strings.TrimSpace(content)).ScrollToBeginning()
}

func (a *App) wrapText(str string, width int) string {
	if len(str) <= width {
		return str
	}

	var result strings.Builder
	lines := strings.Split(str, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line + "\n")
			continue
		}

		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			if len(currentLine)+len(word)+1 <= width {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			} else {
				if currentLine != "" {
					result.WriteString(currentLine + "\n")
				}
				currentLine = word
			}
		}

		if currentLine != "" {
			result.WriteString(currentLine + "\n")
		}
	}

	return result.String()
}

func (a *App) openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		log.Printf("Unsupported os: %s", runtime.GOOS)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

func (a *App) updateStatus(message string) {
	nextCheck := ""
	if a.autoRefresh {
		nextCheck = fmt.Sprintf(" | Next check at: %s", time.Now().Add(5*time.Minute).Format("15:04"))
	}

	nextCheck += "\n[yellow]Ctrl+R[white]: Refresh List | [yellow]Ctrl+O[white]: Open in browser"

	statusText := fmt.Sprintf("[lime]RSS Reader[white] | %s%s\n", message, nextCheck)
	a.status.SetText(statusText)
}

func (a *App) Run() error {
	defer a.database.Close()
	defer a.stopRefresh()
	a.getList(false)
	return a.app.Run()
}
