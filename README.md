# RSS Reader

> A terminal-based RSS news aggregator with multi-source support, content extraction, local storage, and intelligent overview features (coming soon).

[![license](https://img.shields.io/github/license/pardnchiu/rss-reader)](LICENSE)
[![version](https://img.shields.io/github/v/tag/pardnchiu/rss-reader)](https://github.com/pardnchiu/rss-reader/releases)
[![readme](https://img.shields.io/badge/readme-繁體中文-blue)](README.zh.md) 

## Key Features

### Custom Multi-Source News Aggregation
Dynamic RSS feed management with user-friendly source addition, automatic full article content extraction from news websites using goquery, filtering ads and irrelevant elements, providing reader-mode-like text reading and overview, SQLite database storage for offline browsing and fast loading

### Terminal User Interface
Modern TUI interface based on tview with keyboard navigation and real-time updates

### LLM Smart Overview (Coming Soon)
AI-generated 24-hour news overview and article summaries

## Dependencies

- [`github.com/rivo/tview`](https://github.com/rivo/tview) - Terminal UI framework
- [`github.com/gdamore/tcell/v2`](https://github.com/gdamore/tcell/v2) - Terminal event handling
- [`github.com/PuerkitoBio/goquery`](https://github.com/PuerkitoBio/goquery) - HTML content parsing
- [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3) - SQLite database driver

## Usage Guide

### Hotkeys
- `Tab` - Switch between interface panels
- `Ctrl+R` - Manually refresh news
- `Ctrl+O` - Open current news in default browser
- `↑/↓` - Browse news list
- `Enter` - Execute command

### Commands
```bash
# Add RSS feed
add https://example.com/rss.xml

# Remove RSS feed
remove https://example.com/rss.xml
rm https://example.com/rss.xml

# List all feeds
list
```

## Coming Soon

### LLM Smart Overview
- **24-Hour News Summary**: AI-generated daily important news overview
- **Trend Analysis**: Identify trending topics and news patterns

## RSS Recommendations

### BBC News (English)
| Category | Link |
|----------|------|
| General News | https://feeds.bbci.co.uk/news/rss.xml |
| Business | https://feeds.bbci.co.uk/news/business/rss.xml |
| Entertainment & Arts | https://feeds.bbci.co.uk/news/entertainment_and_arts/rss.xml |
| Health | https://feeds.bbci.co.uk/news/health/rss.xml |
| Science & Environment | https://feeds.bbci.co.uk/news/science_and_environment/rss.xml |
| Technology | https://feeds.bbci.co.uk/news/technology/rss.xml |
| World News | https://feeds.bbci.co.uk/news/world/rss.xml |
| BBC Chinese | https://feeds.bbci.co.uk/zhongwen/trad/rss.xml |

### The Guardian News (English)
| Category | Link |
|----------|------|
| World News | https://www.theguardian.com/world/rss |
| Science | https://www.theguardian.com/science/rss |
| Politics | https://www.theguardian.com/politics/rss |
| Business | https://www.theguardian.com/uk/business/rss |
| Technology | https://www.theguardian.com/uk/technology/rss |
| Environment | https://www.theguardian.com/uk/environment/rss |
| Money | https://www.theguardian.com/uk/money/rss |

### Taiwan News Media (Traditional Chinese)
| Media | Link |
|-------|------|
| Liberty Times | https://news.ltn.com.tw/rss/all.xml |
| United Daily News | https://udn.com/rssfeed/news/2/6638?ch=news |
| ETtoday News | https://feeds.feedburner.com/ettoday/news |
| Apple Daily Taiwan | https://tw.appledaily.com/rss |

## License

This project is licensed under the [MIT](LICENSE) License.

## Author

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)