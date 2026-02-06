package app

type ViewState int

const (
	Initializing = iota
	StacksList
	ServicesList
	TaskList
	ContainerAttaching
	ContainerAttached
	ClusterSelection
	CommandSelection
	CustomCommandInput
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
	case ContainerAttaching:
		return "Container Attaching"
	case ContainerAttached:
		return "Container Attached"
	case ClusterSelection:
		return "Cluster Selection"
	case CommandSelection:
		return "Command Selection"
	case CustomCommandInput:
		return "Custom Command Input"
	default:
		return "Unknown"
	}
}
