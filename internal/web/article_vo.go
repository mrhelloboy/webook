package web

// vo: view object 封装用于展示给用户的数据

type ArticleVO struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Content  string `json:"content"`
	Status   uint8  `json:"status"`
	// 计数
	Author     string `json:"author"`
	ReadCnt    int64  `json:"read_cnt"`
	LikeCnt    int64  `json:"like_cnt"`
	CollectCnt int64  `json:"collect_cnt"`
	// 本人是否点赞、收藏
	Liked     bool   `json:"liked"`
	Collected bool   `json:"collected"`
	Ctime     string `json:"ctime"`
	Utime     string `json:"utime"`
}
