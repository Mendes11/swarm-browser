package models

import "github.com/moby/moby/api/types/swarm"

type Task struct {
	TaskID      string
	Node        Node
	ContainerID string
	Status      swarm.TaskState
}
