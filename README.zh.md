# RSS 閱讀器

> 一個基於終端介面的 RSS 新聞聚合器，支援多新聞源、內容萃取、本地儲存和智慧概覽（後續添加）功能。

[![license](https://img.shields.io/github/license/pardnchiu/rss-reader)](LICENSE)
[![version](https://img.shields.io/github/v/tag/pardnchiu/rss-reader)](https://github.com/pardnchiu/rss-reader/releases)
[![readme](https://img.shields.io/badge/readme-English-blue)](README.md) 

## 主要特色

### 自訂多源新聞聚合
動態管理 RSS 訂閱源，支援使用者自由新增任何 RSS 源，使用 goquery 自動從新聞網頁萃取完整文章內容，過濾廣告和無關元素，提供類似閱讀模式的文字讀取與概覽，SQLite 資料庫儲存新聞內容，支援離線瀏覽和快速載入

### 終端使用者介面
基於 tview 的現代化 TUI 介面，支援鍵盤操作和即時更新

### LLM 智慧概覽 (即將推出)
透過 AI 生成 24 小時最新資訊概覽和新聞摘要

## 依賴套件

- [`github.com/rivo/tview`](https://github.com/rivo/tview) - 終端使用者介面框架
- [`github.com/gdamore/tcell/v2`](https://github.com/gdamore/tcell/v2) - 終端事件處理
- [`github.com/PuerkitoBio/goquery`](https://github.com/PuerkitoBio/goquery) - HTML 內容解析
- [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3) - SQLite 資料庫驅動

## 操作指南

### 快捷鍵
- `Tab` - 切換介面區塊
- `Ctrl+R` - 手動更新新聞
- `Ctrl+O` - 在預設瀏覽器中開啟當前新聞
- `↑/↓` - 瀏覽新聞列表
- `Enter` - 執行指令

### 指令
```bash
# 新增 RSS 訂閱源
add https://example.com/rss.xml

# 移除 RSS 訂閱源
remove https://example.com/rss.xml
rm https://example.com/rss.xml

# 列出所有訂閱源
list
```

## 即將推出

### LLM 智慧概覽
- **24小時新聞摘要**: AI 生成當日重要新聞概覽
- **趨勢分析**: 識別熱門話題和新聞趨勢

## RSS 推薦
## RSS 新聞源列表

### BBC 新聞 (英文)
| 分類 | 連結 |
|------|----------|
| 綜合新聞 | https://feeds.bbci.co.uk/news/rss.xml |
| 商業財經 | https://feeds.bbci.co.uk/news/business/rss.xml |
| 娛樂藝術 | https://feeds.bbci.co.uk/news/entertainment_and_arts/rss.xml |
| 健康醫療 | https://feeds.bbci.co.uk/news/health/rss.xml |
| 科學環境 | https://feeds.bbci.co.uk/news/science_and_environment/rss.xml |
| 科技資訊 | https://feeds.bbci.co.uk/news/technology/rss.xml |
| 國際新聞 | https://feeds.bbci.co.uk/news/world/rss.xml |
| BBC中文 | https://feeds.bbci.co.uk/zhongwen/trad/rss.xml |

### The Guardian 新聞 (英文)
| 分類 | 連結 |
|------|----------|
| 國際新聞 | https://www.theguardian.com/world/rss |
| 科學資訊 | https://www.theguardian.com/science/rss |
| 政治新聞 | https://www.theguardian.com/politics/rss |
| 商業財經 | https://www.theguardian.com/uk/business/rss |
| 科技資訊 | https://www.theguardian.com/uk/technology/rss |
| 環境議題 | https://www.theguardian.com/uk/environment/rss |
| 理財資訊 | https://www.theguardian.com/uk/money/rss |

### 台灣新聞媒體 (繁體中文)
| 媒體 | 連結 |
|----------|----------|
| 自由時報 | https://news.ltn.com.tw/rss/all.xml |
| 聯合新聞網 | https://udn.com/rssfeed/news/2/6638?ch=news |
| ETtoday 新聞雲 | https://feeds.feedburner.com/ettoday/news |
| 台灣蘋果日報 | https://tw.appledaily.com/rss |

## 授權條款

此專案採用 [MIT](LICENSE) 授權條款。

## 作者

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)