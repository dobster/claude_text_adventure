package main

import (
	"bufio"
	"fmt"
	"os"

	"textadventure/engine"
)

func main() {
	game := engine.NewSession()
	for _, line := range game.FlushOutput() {
		fmt.Println(line)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for game.IsRunning() {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		game.HandleInput(scanner.Text())
		for _, line := range game.FlushOutput() {
			fmt.Println(line)
		}
	}
}
