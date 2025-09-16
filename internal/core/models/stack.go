package models

// Stack represents a Docker Swarm stack with its services
type Stack struct {
	Name string
}

func (s Stack) String() string {
	return s.Name
}
