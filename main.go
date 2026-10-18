package main

import (
	"fmt"
	"os"

	"vyshkat/park"
)

type Container struct {
	inspector park.Inspector
	fleet     *park.Fleet
}

func (c *Container) Inspector() park.Inspector {
	if c.inspector == nil {
		c.inspector = park.ServiceCenter{}
	}
	return c.inspector
}

func (c *Container) Fleet() *park.Fleet {
	if c.fleet == nil {
		c.fleet = park.NewFleet(c.Inspector())
	}
	return c.fleet
}

func main() {
	container := &Container{}
	console := NewConsole(container.Fleet(), os.Stdin, os.Stdout)
	if err := console.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка чтения:", err)
		os.Exit(1)
	}
}
