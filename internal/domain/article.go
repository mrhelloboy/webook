package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

// Abstract 返回文章摘要
func (a Article) Abstract() string {
	// 考虑中文问题
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	return string(cs[:100])
}

type Author struct {
	Id   int64
	Name string
}

type ArticleStatus uint8

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

// ToUint8 converts the status to uint8.
func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

// NonPublished returns true if the status is not published.
func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

// String returns the string representation of the status.
func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusPrivate:
		return "private"
	case ArticleStatusUnpublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	default:
		return "unknown"
	}
}
