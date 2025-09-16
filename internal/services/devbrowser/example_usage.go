// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/mendes11/swarm-browser/internal/services/devbrowser"
)

// Example usage of the DevBrowser
func main() {
	// Option 1: Load configuration from dev-clusters.yaml file
	browser1, err := devbrowser.New("dev-local", "dev-clusters.yaml")
	if err != nil {
		log.Fatalf("Failed to create browser with config file: %v", err)
	}
	defer browser1.Close()

	// Option 2: Create custom configuration programmatically
	customConfig := &devbrowser.DevConfig{
		Clusters: map[string]models.Cluster{
			"my-dev": {
				Name: "My Development Cluster",
				Host: "localhost",
				Nodes: map[string]models.Node{
					"manager": {Host: "localhost", Hostname: "dev-manager"},
					"worker1": {Host: "localhost", Hostname: "dev-worker-1"},
					"worker2": {Host: "localhost", Hostname: "dev-worker-2"},
				},
			},
		},
		Stacks: []devbrowser.StackConfig{
			{
				Name:        "my-app",
				ClusterName: "my-dev",
				Services: []devbrowser.ServiceConfig{
					{
						Name:         "api",
						DesiredTasks: 3,
						RunningTasks: 3,
					},
					{
						Name:         "database",
						DesiredTasks: 1,
						RunningTasks: 1,
						Tasks: []devbrowser.TaskConfig{
							{
								NodeName: "manager",
								Status:   "running",
							},
						},
					},
				},
			},
		},
	}

	browser2, err := devbrowser.NewWithConfig("my-dev", customConfig)
	if err != nil {
		log.Fatalf("Failed to create browser with custom config: %v", err)
	}
	defer browser2.Close()

	// Use the browser to explore the mock cluster
	ctx := context.Background()

	// Get cluster information
	cluster := browser2.GetCluster()
	fmt.Printf("Connected to cluster: %s (%s)\n", cluster.Name, cluster.Host)
	fmt.Printf("Nodes in cluster:\n")
	for nodeName, node := range cluster.Nodes {
		fmt.Printf("  - %s: %s (%s)\n", nodeName, node.Hostname, node.Host)
	}

	// List stacks
	stacks, err := browser2.ListStacks(ctx)
	if err != nil {
		log.Fatalf("Failed to list stacks: %v", err)
	}

	fmt.Println("\nAvailable stacks:")
	for _, stack := range stacks {
		fmt.Printf("  Stack: %s\n", stack.Name)

		// List services in each stack
		services, err := browser2.ListServices(ctx, stack)
		if err != nil {
			log.Printf("Failed to list services for stack %s: %v", stack.Name, err)
			continue
		}

		for _, service := range services {
			fmt.Printf("    Service: %s (%d/%d replicas)\n",
				service.Name, service.RunningTasks, service.DesiredTasks)

			// List tasks for each service
			tasks, err := browser2.ListTasks(ctx, service)
			if err != nil {
				log.Printf("Failed to list tasks for service %s: %v", service.Name, err)
				continue
			}

			for _, task := range tasks {
				fmt.Printf("      Task %s on %s: %v\n",
					task.TaskID, task.Node.Hostname, task.Status)
			}
		}
	}

	// Attach to a service (example - would be interactive in real usage)
	if len(stacks) > 0 {
		firstStack := stacks[0]
		services, err := browser2.ListServices(ctx, firstStack)
		if err == nil && len(services) > 0 {
			firstService := services[0]
			fmt.Printf("\nAttaching to service %s...\n", firstService.Name)

			// This would normally be used with a terminal UI
			// Here we use echo to demonstrate the connection
			conn, err := browser2.AttachToService(ctx, firstService, []string{"echo", "Hello from mock container!"})
			if err != nil {
				log.Printf("Failed to attach to service: %v", err)
			} else {
				defer conn.Close()
				fmt.Println("Successfully attached to service (connection established)")
			}
		}
	}

	// Example: Using multiple clusters in the same config
	fmt.Println("\n--- Switching to dev-staging cluster ---")
	browser3, err := devbrowser.New("dev-staging", "dev-clusters.yaml")
	if err != nil {
		log.Printf("Failed to connect to dev-staging: %v", err)
	} else {
		defer browser3.Close()

		stacks, err := browser3.ListStacks(ctx)
		if err == nil {
			fmt.Printf("Dev-staging cluster has %d stacks\n", len(stacks))
			for _, stack := range stacks {
				fmt.Printf("  - %s\n", stack.Name)
			}
		}
	}
}