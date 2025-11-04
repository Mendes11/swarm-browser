# V1
First version is a simple TUI that allows you to drill-down stacks -> services -> containers

The user must be able to:
 * Go back
 * Drill Down into a Stack, displaying all services, and tasks
 * [A]ttach to a containers of a service
 * [A]ttach to a container of a task
 * Filter Stacks and Services by name

## Tasks

[ ] Custom hooks for switching cluster -- Support custom commands when switching to a cluster (ie: switch vpn)

[ ] What's the ideal way to distribute clusters.yml? Have it stored in an S3 bucket and just download from it as part of the install script?

[ ] How do I distribute this?

[ ] Create a CLI on the core models

[ ] Create a CLI to attach to a service

[ ] Implement switch between clusters

[ ] Come-up with a name
