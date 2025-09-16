package models

import (
	"fmt"
)

type Service struct {
	ID           string
	Name         string
	RunningTasks uint64
	DesiredTasks uint64
	Stack        Stack
}

func (s Service) String() string {
	return fmt.Sprintf("%s (%d/%d replicas)", s.Name, s.RunningTasks, s.DesiredTasks)
}
