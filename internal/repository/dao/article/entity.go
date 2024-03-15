package article

// Article 制作库
type Article struct {
	Id       int64  `gorm:"primary_key,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(1024)" bson:"title,omitempty"`
	Content  string `gorm:"type=BLOB" bson:"content,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}

// PublishedArticle 线上库，表结构跟制作库一致，表示已发表的状态
type PublishedArticle struct {
	Article `bson:"inline"`
}
