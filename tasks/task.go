package tasks

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func NewTask(title string) Task {
	return Task{Title: title}
}
