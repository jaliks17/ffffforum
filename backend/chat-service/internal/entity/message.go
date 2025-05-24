package entity

type Message struct {
	ID      string `json:"id"`
	User    string `json:"user"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
}
