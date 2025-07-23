package model

type Task struct {
	ID         string
	Links      map[string]struct{}
	Downloaded map[string]string
}
