package app

type ViewState int

const (
	Initializing = iota
	StacksList
	ServicesList
	TaskList
	ContainerAttached
)
