package model

type Task struct {
	ID         string
	Links      []string
	Downloaded map[string]string
}
