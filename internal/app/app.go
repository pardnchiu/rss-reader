package app

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"rss-reader/internal/api"
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

	a.llmView = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)
	a.llmView.SetBorder(true).
		SetTitle("Summary").
		SetTitleAlign(tview.AlignLeft)
	summary, _ := a.database.GetKey("summary")
	if summary != "" {
		a.llmView.SetText(summary).ScrollToBeginning()
	}

	leftView := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.llmView, 0, 1, true).
		AddItem(a.list, 18, 0, true)

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
				a.app.SetFocus(a.llmView)
			} else {
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
		if len(parts) < 2 {
			a.showCommand("add [URL]")
			return
		}
		url := parts[1]
		a.collector.Add(url)
		a.showCommand(fmt.Sprintf("Add RSS: %s", cmd))
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

	case "apikey":
		if len(parts) < 2 {
			a.showCommand("apikey [API_KEY]")
			return
		}
		key := parts[1]
		if err := a.database.SetKey("apikey", key); err != nil {
			a.showCommand(fmt.Sprintf("Failed to set API key: %v", err))
			return
		}
		a.showCommand("API key set successfully.")

	case "config":
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

	key, _ := a.database.GetKey("apikey")
	if key == "" {
		key = "Not set"
	}

	result := fmt.Sprintf("API Key: %s\n\n%s\n", key, "RSS feed list:")
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
		// 1. 如果列表為空，先從資料庫載入
		if len(a.articles) < 1 {
			storedArticles, err := a.database.Get(72)
			if err == nil && len(storedArticles) > 0 {
				a.app.QueueUpdateDraw(func() {
					a.articles = storedArticles
					a.filteredArticles = a.articles
					a.updateList()
					a.updateStatus(fmt.Sprintf("Get %d news from Database", len(a.articles)))
				})
			}
		}

		// 2. 從 RSS 獲取新文章
		newArticles, err := a.collector.GetNews()
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.updateStatus(fmt.Sprintf("Failed to get news list: %v", err))
			})
			return
		}

		// 3. 檢查並插入資料庫中沒有的文章
		newCount := 0
		finalArticles := make([]model.News, 0)

		for _, article := range newArticles {
			stored, err := a.database.GetFromURL(article.URL)
			if err != nil {
				// 資料庫沒有這篇文章，計入新文章
				newCount++
				finalArticles = append(finalArticles, article)
			} else {
				// 使用資料庫中的發布時間
				article.PublishedAt = stored.PublishedAt
				finalArticles = append(finalArticles, article)
			}
		}

		// 按發布時間排序 (新到舊)
		sort.Slice(finalArticles, func(i, j int) bool {
			return finalArticles[i].PublishedAt.After(finalArticles[j].PublishedAt)
		})

		// 4. 非同步載入新文章的完整內容
		if newCount > 0 {
			go a.loadContent(finalArticles)
		}

		// 5. 更新 UI
		a.app.QueueUpdateDraw(func() {
			a.articles = finalArticles
			a.filteredArticles = finalArticles
			a.updateList()
			if newCount > 0 {
				a.updateStatus(fmt.Sprintf("Found %d new articles, getting full content...", newCount))
			} else {
				a.updateStatus("All news are up to date.")
			}
		})
	}()
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

		key, _ := a.database.GetKey("apikey")
		if key != "" {
			api.ApiKey = key
		}
		summary, _ := a.database.GetKey("summary")
		systemPrompt := strings.TrimSpace(
			fmt.Sprintf(`=== 系統資訊 ===
當前時間：%s
作業系統與執行環境：%s

=== 指令說明 ===
你是專業的新聞概要整理助手。請根據輸入的新聞內容（標題、日期、完整內容）提取重點，生成結構化的本日概要並分析趨勢。

=== 輸出格式要求 ===
使用繁體中文，按以下順序組織內容：

## 重大要聞
- 國際局勢、政府政策、重大社會事件
- 保留詳細資訊，避免過度壓縮
- 標註時間與影響程度

## 科技與金融
- 科技趨勢、市場動向、經濟指標
- 重點關注創新技術與投資機會
- 標註相關股市或產業影響

## 生活資訊
- 天氣預報、交通狀況、消費資訊
- 健康醫療、教育文化相關消息

## 趨勢分析
- 對比前次概要，標註變化趨勢
- 識別持續發展或新興話題
- 預測可能後續發展

=== 處理原則 ===
1. 盡可能地保留舊有概要內容（包含24小時內、重大消息）
2. 依新聞重要性與時效性排序
3. 盡可能從相近新聞中補充細節
4. 保留數據、時間、人名等關鍵細節
5. 標註消息來源可信度
6. 突出與前次概要的差異變化


=== 前次概要 ===
%s`,
				time.Now().Format("2006年01月02日 15:04:05"),
				runtime.GOOS+"/"+runtime.GOARCH,
				summary,
			),
		)
		messages := []api.Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
		}

		if summary == "" {
			arr, _ := a.database.Get(24)
			sort.Slice(arr, func(i, j int) bool {
				return arr[i].PublishedAt.After(arr[j].PublishedAt)
			})

			for _, item := range arr {
				messages = append(messages, api.Message{
					Role:    "user",
					Content: fmt.Sprintf("Title: %s\nSource: %s\nPublishedAt: %s\nContent: %s", item.Title, item.Source, item.PublishedAt.Local().Format("2006-01-02 15:04"), item.Content),
				})
			}
		}

		for _, item := range news {
			messages = append(messages, api.Message{
				Role:    "user",
				Content: fmt.Sprintf("Title: %s\nSource: %s\nPublishedAt: %s\nContent: %s", item.Title, item.Source, item.PublishedAt.Local().Format("2006-01-02 15:04"), item.Content),
			})
		}

		summary, err := api.AskWithSmallModel(messages)
		if err != nil {
			a.llmView.SetText(err.Error()).ScrollToBeginning()
			return
		}
		a.database.SetKey("summary", summary)
		a.llmView.SetText(summary).ScrollToBeginning()
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
