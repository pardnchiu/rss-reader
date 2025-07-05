package model

import "time"

type News struct {
	Title       string
	Content     string
	Source      string
	URL         string
	PublishedAt time.Time

	FullContent *string
	Author      *string
	WordCount   *int
}

type NewsContent struct {
	Title     string
	Author    string
	Content   string
	WordCount int
}
