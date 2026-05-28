package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func Container(code string) (string, string) {
	ctx := context.Background()
	os.Mkdir("files", 0777)
	file, err := os.Create("files/scene.py")
	if err != nil {
		fmt.Println("File error: ", err)
		panic(err)
	}

	defer file.Close()
	file.WriteString(code)

	dockerCli, err := client.New(client.FromEnv)

	if err != nil {
		fmt.Println("Docker error: ", err)
		panic(err)
	}

	defer dockerCli.Close()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("CWD error: ", err)
		panic(err)
	}

	bindPath := fmt.Sprintf("%s/files:/app", cwd)

	resp, err := dockerCli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: "manim-runner",
		Config: &container.Config{
			Cmd: []string{"manim", "-pql", "/app/scene.py", "SceneName"},
			Tty: false,
		},
		HostConfig: &container.HostConfig{
			Binds: []string{bindPath},
		},
	})
	if err != nil {
		fmt.Println("Container create error: ", err)
		panic(err)
	}

	if _, err := dockerCli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		fmt.Println("Container start error: ", err)
		panic(err)
	}

	wait := dockerCli.ContainerWait(ctx, resp.ID, client.ContainerWaitOptions{})
	select {
	case err := <-wait.Error:
		if err != nil {
			panic(err)
		}
	case <-wait.Result:
	}

	out, err := dockerCli.ContainerLogs(ctx, resp.ID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if out == nil {
		panic("out is nil")
	}

	_, err = stdcopy.StdCopy(&stdout, &stderr, out)
	if err != nil {
		panic("out: " + err.Error())
	}

	dockerCli.ContainerRemove(ctx, resp.ID, client.ContainerRemoveOptions{Force: true})

	return stderr.String(), stdout.String()
}
