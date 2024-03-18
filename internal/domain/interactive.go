package domain

type Interactive struct {
	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`
	Liked      bool  `json:"liked"`
	Collected  bool  `json:"collected"`
}

type Self struct {
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type Collection struct {
	Name  string
	Uid   int64
	Items []Resource
}
