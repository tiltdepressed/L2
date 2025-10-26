package parser

import (
	"15/service"
	"15/shell"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Command struct {
	Name   string
	Args   []string
	Input  io.Reader
	Output io.Writer
	Cmd    *exec.Cmd
}

func ExecutePipeline(commands []*Command, shell *shell.Shell) error {
	if len(commands) == 0 {
		return nil
	}

	// Создаем пайпы между командами
	pipes := make([]*io.PipeWriter, len(commands)-1)
	for i := 0; i < len(commands)-1; i++ {
		reader, writer := io.Pipe()
		commands[i].Output = writer
		commands[i+1].Input = reader
		pipes[i] = writer
	}

	// Устанавливаем граничные потоки
	if commands[0].Input == nil {
		commands[0].Input = os.Stdin
	}
	if commands[len(commands)-1].Output == nil {
		commands[len(commands)-1].Output = os.Stdout
	}

	// Запускаем команды
	var wg sync.WaitGroup
	errCh := make(chan error, len(commands))

	for i, cmd := range commands {
		wg.Add(1)
		go func(i int, cmd *Command) {
			defer wg.Done()

			// Закрываем пайп после записи
			if i < len(commands)-1 {
				if writer, ok := cmd.Output.(*io.PipeWriter); ok {
					defer writer.Close()
				}
			}

			err := ExecuteCommand(cmd, shell)
			if err != nil {
				errCh <- err
			}
		}(i, cmd)
	}

	wg.Wait()
	close(errCh)

	// Вернуть первую ошибку
	for err := range errCh {
		return err
	}
	return nil
}

func ExecuteCommand(cmd *Command, shell *shell.Shell) error {
	if isBuiltin(cmd.Name) {
		return ExecuteBuiltin(cmd)
	}
	return ExecuteExternal(cmd, shell)
}

func isBuiltin(name string) bool {
	return name == "cd" || name == "pwd" || name == "echo" || name == "kill" || name == "ps"
}

func ParseCommand(line string) *Command {
	if line == "" {
		return nil
	}

	parts := strings.Fields(line)

	cmd := &Command{
		Name: parts[0],
		Args: parts[1:],
	}
	return cmd

}

func ExecuteBuiltin(cmd *Command) error {
	switch cmd.Name {
	case "cd":
		return service.Cd(cmd.Args[0])
	case "pwd":
		return service.Pwd(cmd.Output)
	case "echo":
		return service.Echo(cmd.Args, cmd.Output)
	case "kill":
		id, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		return service.Kill(id)
	case "ps":
		return service.Ps(cmd.Output)
	default:
		return nil
	}
}

func ExecuteExternal(cmd *Command, shell *shell.Shell) error {
	c := exec.Command(cmd.Name, cmd.Args...)
	if cmd.Input != nil {
		c.Stdin = cmd.Input
	} else {
		c.Stdin = os.Stdin
	}
	if cmd.Output != nil {
		c.Stdout = cmd.Output
	} else {
		c.Stdout = os.Stdout
	}

	c.Stderr = os.Stderr
	shell.AddProcess(c)

	err := c.Run()
	shell.RemoveProcess(c)

	return err
}
