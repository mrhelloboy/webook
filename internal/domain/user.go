package domain

import "time"

type User struct {
	Id         int64
	Email      string
	Password   string
	Phone      string
	Nickname   string
	WechatInfo WechatInfo
	Ctime      time.Time
}
