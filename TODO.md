# V1
First version is a simple TUI that allows you to drill-down stacks -> services -> containers

The user must be able to:
 * Go back
 * Drill Down into a Stack, displaying all services
 * [A]ttach to all containers of a service
 * [E]nter into a single container of a service (picked by first successful connection)
   * Prompt the user which bash command to execute (default to bash)

## Tasks

- [ ] Add attach option to Service level
    - [ ] Keybind: a
    - [ ] Attach to ALL running containers, and print into a dedicated container

- [ ] Add exec option to Service Level
    - [ ] Asks for default bash command (defaults to bash)
    - [ ] Create new Exec and attach. This first version should takeover the entire TUI (terminal RawMode)

- [ ] Add Container Listing
- [ ] Add Attach and Exec to it