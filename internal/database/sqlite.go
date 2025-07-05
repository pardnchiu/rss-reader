package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	"rss-reader/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite() (*SQLite, error) {
	var dbPath string

	if customPath := os.Getenv("RSS_DB_PATH"); customPath != "" {
		dbPath = customPath
	} else {
		if isDev() {
			// 開發環境：使用工作目錄
			wd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			dbPath = filepath.Join(wd, "rss.db")
		} else {
			// 生產環境：使用執行檔目錄或用戶目錄
			path, err := os.Executable()
			if err != nil {
				return nil, err
			}
			dir := filepath.Dir(path)
			dbPath = filepath.Join(dir, "rss.db")

			if !canWrite(dir) {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return nil, err
				}
				dbPath = filepath.Join(homeDir, ".rss-reader", "rss.db")
				if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
					return nil, err
				}
			}
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &SQLite{db: db}
	if err := s.create(); err != nil {
		return nil, err
	}

	return s, nil
}

func isDev() bool {
	if _, err := os.Stat("go.mod"); err == nil {
		return true
	}
	return os.Getenv("GO_ENV") == "development"
}

func canWrite(dir string) bool {
	testFile := filepath.Join(dir, ".write_test")
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		os.Remove(testFile)
		return true
	}
	return false
}

func (s *SQLite) create() error {
	query := `
    CREATE TABLE IF NOT EXISTS news (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        url TEXT UNIQUE NOT NULL,
        content TEXT,
        full_content TEXT,
        source TEXT,
        author TEXT,
        word_count INTEGER,
        published_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

		CREATE TABLE IF NOT EXISTS feeds (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT UNIQUE NOT NULL,
        dismiss INTEGER DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

		CREATE TABLE IF NOT EXISTS data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        key TEXT UNIQUE NOT NULL,
        value TEXT
    );

    CREATE INDEX IF NOT EXISTS idx_news_url ON news(url);
    CREATE INDEX IF NOT EXISTS idx_news_published_at ON news(published_at);
    CREATE INDEX IF NOT EXISTS idx_news_source ON news(source);
    CREATE INDEX IF NOT EXISTS idx_feeds_url ON feeds(url);
    CREATE INDEX IF NOT EXISTS idx_feeds_dismiss ON feeds(dismiss);
    CREATE INDEX IF NOT EXISTS idx_data_key ON data(key);
    `

	_, err := s.db.Exec(query)
	return err
}

func (s *SQLite) Insert(news model.News, content *model.NewsContent) error {
	query := `
	INSERT OR REPLACE INTO news (
		title, 
		url, 
		content, 
		full_content, 
		source, 
		author, 
		word_count, 
		published_at
	)
  VALUES (
		?, 
		?, 
		?, 
		?, 
		?, 
		?, 
		?, 
		?
	)`

	fullContent := ""
	author := ""
	wordCount := 0

	if content != nil {
		fullContent = strings.TrimSpace(content.Content)
		author = strings.TrimSpace(content.Author)
		wordCount = content.WordCount
	}

	_, err := s.db.Exec(query,
		strings.TrimSpace(news.Title),
		strings.TrimSpace(news.URL),
		strings.TrimSpace(news.Content),
		fullContent,
		strings.TrimSpace(news.Source),
		author,
		wordCount,
		news.PublishedAt,
	)

	return err
}

func (s *SQLite) Get(hours int) ([]model.News, error) {
	query := `
	SELECT title, url, content, full_content, source, author, word_count, published_at
	FROM news 
	WHERE published_at >= datetime('now', '-' || ? || ' hours')
	ORDER BY published_at DESC`

	result, err := s.db.Query(query, hours)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var arr []model.News

	for result.Next() {
		var article model.News
		var fullContent, author string
		var wordCount int

		err := result.Scan(
			&article.Title,
			&article.URL,
			&article.Content,
			&fullContent,
			&article.Source,
			&author,
			&wordCount,
			&article.PublishedAt,
		)

		if err != nil {
			continue
		}

		if fullContent != "" {
			article.FullContent = &fullContent
		}
		if author != "" {
			article.Author = &author
		}
		if wordCount > 0 {
			article.WordCount = &wordCount
		}

		arr = append(arr, article)
	}

	return arr, nil
}

func (s *SQLite) GetFromURL(url string) (*model.News, error) {
	query := `
	SELECT title, url, content, full_content, source, author, word_count, published_at 
	FROM news 
	WHERE url = ?`

	var stored model.News
	var fullContent, author string
	var wordCount int

	err := s.db.QueryRow(query, url).Scan(
		&stored.Title,
		&stored.URL,
		&stored.Content,
		&fullContent,
		&stored.Source,
		&author,
		&wordCount,
		&stored.PublishedAt,
	)

	if err != nil {
		return nil, err
	}
	if fullContent != "" {
		stored.FullContent = &fullContent
	}
	if author != "" {
		stored.Author = &author
	}
	if wordCount > 0 {
		stored.WordCount = &wordCount
	}

	return &stored, nil
}

func (s *SQLite) InsertFeed(url string) error {
	query := `
	INSERT OR REPLACE INTO feeds (
		url, 
		dismiss, 
		updated_at
	)
	VALUES (
		?, 
		0,
		CURRENT_TIMESTAMP
	)`

	_, err := s.db.Exec(query, strings.TrimSpace(url))
	return err
}

func (s *SQLite) RemoveFeed(url string) error {
	query := `
	UPDATE feeds 
	SET 
		dismiss = 1, 
		updated_at = CURRENT_TIMESTAMP
	WHERE url = ?`

	_, err := s.db.Exec(query, strings.TrimSpace(url))
	return err
}

func (s *SQLite) GetFeed() ([]string, error) {
	query := `
	SELECT url
	FROM feeds 
	WHERE dismiss = 0 
	ORDER BY created_at ASC`

	result, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var feeds []string
	for result.Next() {
		var url string
		if err := result.Scan(&url); err != nil {
			continue
		}
		feeds = append(feeds, url)
	}

	return feeds, nil
}

func (s *SQLite) GetKey(key string) (string, error) {
	query := `
	SELECT value
	FROM data 
	WHERE key = ?`

	var value string
	err := s.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (s *SQLite) SetKey(key, value string) error {
	query := `
	INSERT OR REPLACE INTO data (
		key, 
		value
	)
	VALUES (
		?, 
		?
	)`

	_, err := s.db.Exec(query, strings.TrimSpace(key), strings.TrimSpace(value))
	return err
}

func (s *SQLite) Close() error {
	return s.db.Close()
}
