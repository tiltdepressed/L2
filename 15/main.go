package main

import (
	"15/parser"
	"15/shell"
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	shell := &shell.Shell{}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		for sig := range sigCh { // Ctrl+C
			if sig == syscall.SIGINT {
				fmt.Println("\nReceived SIGINT (Ctrl+C)")
				shell.KillAllProcesses()
				fmt.Print("$ ")
			}

		}
	}()

	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("$ ")
		if !sc.Scan() {
			fmt.Println("\nReceived EOF (Ctrl+D) - Goodbye!")
			shell.KillAllProcesses()
			break
		}

		input := strings.TrimSpace(sc.Text())
		if input == "" {
			continue
		}

		commandStrings := strings.Split(input, "|")
		var commands []*parser.Command
		for _, cmdStr := range commandStrings {
			cmd := parser.ParseCommand(cmdStr)
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
		if len(commands) == 1 {
			cmd := commands[0]

			if cmd.Input == nil {
				cmd.Input = os.Stdin
			}
			if cmd.Output == nil {
				cmd.Output = os.Stdout
			}
			err := parser.ExecuteCommand(cmd, shell)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		} else {
			parser.ExecutePipeline(commands, shell)
		}
	}
}
