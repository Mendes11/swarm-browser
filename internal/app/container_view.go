package app

// ExitContainerViewMsg is sent when exiting the container session
type ExitContainerViewMsg struct {
	Err error
}
