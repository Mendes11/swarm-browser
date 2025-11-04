# V1
First version is a simple TUI that allows you to drill-down stacks -> services -> containers

The user must be able to:
 * Go back
 * Drill Down into a Stack, displaying all services, and tasks
 * [A]ttach to a containers of a service
 * [A]ttach to a container of a task
 * Filter Stacks and Services by name

## Tasks

[ ] Custom hooks for switching clusters -- Support custom commands when switching to a cluster (ie: switch vpn)

[ ] Implement switch between clusters

[ ] What's the ideal way to distribute clusters.yml? Have it stored in an S3 bucket and just download from it as part of the install script?

[ ] How do I distribute this binary? | Come-up with a name

----

[ ] Create a CLI on the core models

[ ] Create a CLI to attach to a service

[ ] Refactor main model -> Objetive is to reduce complexity, and readability.
