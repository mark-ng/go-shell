package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

const prompt = "🔥 "

var childProcess *os.Process

func handler(s os.Signal) {
	switch s {
	case syscall.SIGINT:
		if childProcess != nil && childProcess.Pid != 0 {
			if err := syscall.Kill(childProcess.Pid, syscall.SIGINT); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to send SIGINT to child: %v\n", err)
			}
		}
	}
}

func parseCommandLine(commandLine string) [][]string {
	commands := strings.Split(commandLine, "|")
	var commandsArr [][]string

	for _, v := range commands {
		tokens := strings.Fields(v)
		var tokensArr []string
		for _, t := range tokens {
			tokensArr = append(tokensArr, t)
		}
		commandsArr = append(commandsArr, tokensArr)
	}

	return commandsArr
}

func main() {
	// Handle SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		for {
			s := <-sigs
			handler(s)
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	// Main Loop
	for {
		fmt.Printf(prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			// Handle EOF
			if err == io.EOF {
				fmt.Println("exit")
				os.Exit(0)
			}
			fmt.Println("Error reading user input")
			os.Exit(1)
		}

		// Parse Input
		commandsArr := parseCommandLine(input)

		// Handle Empty Input
		if len(commandsArr) == 1 && len(commandsArr[0]) == 0 {
			continue
		}
		// Handle Shell builtins
		switch commandsArr[0][0] {
		case "quit":
			os.Exit(0)
		case "help":
			fmt.Println("Shell builtins:")
			fmt.Println(fmt.Sprintf("%-20v - Help", "help"))
			fmt.Println(fmt.Sprintf("%-20v - Kill Shell", "quit, Ctrl + D"))
			continue
		}

		// Handle Program Execution
		childP := exec.Command(commandsArr[0][0], commandsArr[0][1:]...)
		childP.Stdin = os.Stdin
		childP.Stdout = os.Stdout
		childP.Stderr = os.Stderr

		if err := childP.Start(); err != nil {
			fmt.Printf("Failed to start process %v\n", err)
			continue
		}
		if err := childP.Wait(); err != nil {
			fmt.Printf("Command failed: %v\n", err)
			continue
		}

		childProcess = nil
	}
}
