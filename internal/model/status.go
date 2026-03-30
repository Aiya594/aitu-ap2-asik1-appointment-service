package model

type Status string

const (
	StatusNew  Status = "new"
	InProgress Status = "in_progress"
	Done       Status = "done"
)

var validTransitions = map[Status]Status{
	StatusNew:  InProgress,
	InProgress: Done,
}
