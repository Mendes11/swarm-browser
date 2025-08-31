package types

import (
	"fmt"

	"github.com/moby/moby/api/types/swarm"
)

type Service swarm.Service

func (s Service) String() string {
	return fmt.Sprintf("%s (%d/%d replicas)", s.Spec.Name, s.ServiceStatus.RunningTasks, s.ServiceStatus.DesiredTasks)
}
