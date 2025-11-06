package app

type ViewState int

const (
	Initializing = iota
	StacksList
	ServicesList
	TaskList
	ContainerAttached
	ClusterSelection
)

func (v ViewState) String() string {
	switch v {
	case Initializing:
		return "Initializing"
	case StacksList:
		return "Stacks List"
	case ServicesList:
		return "Services List"
	case TaskList:
		return "Task List"
	case ContainerAttached:
		return "Container Attached"
	case ClusterSelection:
		return "Cluster Selection"
	default:
		return "Unknown"
	}
}
