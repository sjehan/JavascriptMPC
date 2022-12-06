package main

import (
	builder "ixxoprivacy/pkg/builder"
	"ixxoprivacy/pkg/garbler"
	"ixxoprivacy/pkg/runner"
	"log"
	"os"

	"github.com/mitchellh/cli"
)

type buildCommand struct{}
type runCommand struct{}
type garbleCommand struct{}

func (c *buildCommand) Help() string {
	return "This command builds a circuit from a javascript file. Note that the Javascript has specific conventions for MPC, refer to the documentation."
}
func (c *buildCommand) Run(args []string) int {
	if len(args) != 1 {
		log.Println("You have to provide the name of the file to build")
		return 1
	}
	//var filename := args[0]
	fileName := args[0]
	builder.BuildCircuit(fileName)
	return 0
}
func (c *buildCommand) Synopsis() string {
	return "This command builds a circuit from a javascript file. Note that the Javascript has specific conventions for MPC, refer to the documentation."

}

func (c *runCommand) Help() string {
	return "Runs a circuit. You have to provide the compiled circuit file as a parameter as the first argument, and the list of imput"
}
func (c *runCommand) Run(args []string) int {
	if len(args) == 0 {
		log.Println("You have to provide the name of the file to run")
		return 1
	}
	compiledCircuit := args[0]
	inputFiles := args[1:]
	runner.RunCircuit(compiledCircuit, inputFiles)
	return 0
}

func (c *runCommand) Synopsis() string {
	return "Garbles a circuit. Add true as a second argument to see the garbled circuit in debug mode."
}

func (c *garbleCommand) Help() string {
	return "Garbles a circuit. Add true as a second argument to see the garbled circuit in debug mode."
}
func (c *garbleCommand) Run(args []string) int {
	if len(args) == 0 {
		log.Println("You have to provide the name of the file to garble")
		return 1
	}
	if len(args) == 1 {
		garbler.GarbleCompiledCircuit(args[0], false, 8)
	} else {
		if args[1] == "true" {
			garbler.GarbleCompiledCircuit(args[0], true, 8)
		} else if args[1] == "false" {
			garbler.GarbleCompiledCircuit(args[0], false, 8)
		} else {
			log.Println("Second argument must be true or false")
			return 1
		}
	}
	return 0
}
func (c *garbleCommand) Synopsis() string {
	return "Garbles a circuit"
}

func main() {
	c := cli.NewCLI("rockengine", "0.0.1")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &buildCommand{}, nil
		},
		"run": func() (cli.Command, error) {
			return &runCommand{}, nil
		},
		"garble": func() (cli.Command, error) {
			return &garbleCommand{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
