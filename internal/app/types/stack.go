package types

// Stack represents a Docker Swarm stack with its services
type Stack struct {
	Name     string
	Services []Service
}

func (s Stack) String() string {
	return s.Name
}
