package dto

type ConnectionRequest struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

type FileRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
