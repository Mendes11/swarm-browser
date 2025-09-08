package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/mendes11/swarm-browser/internal/services/connector"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/term"
)

func main() {
	connector, err := connector.New()
	if err != nil {
		panic(err)
	}
	defer connector.Close()

	cli, err := connector.ClientForHost("worker-02.janushcp.com")
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	execResp, err := cli.ContainerExecCreate(context.Background(), "ac43c88fac17", container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"bash"},
	})
	if err != nil {
		panic(err)
	}

	log.Println("Container Exec Started")
	container, err := cli.ContainerExecAttach(context.Background(), execResp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		panic(err)
	}
	defer container.Close()
	// Set Terminal to Raw Mode
	oldState, err := term.MakeRaw(os.Stdin.Fd())
	if err != nil {
		panic(err)
	}
	defer term.RestoreTerminal(os.Stdin.Fd(), oldState)
	log.Println("Connected to Container")
	go io.Copy(container.Conn, os.Stdout)
	io.Copy(os.Stdin, container.Conn)
}
