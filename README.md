# Graph / Topological Sort

This is a small library I built to scratch a personal itch.

## What is topological sorting?

A directed acyclic graph (DAG) is often used to represent dependencies between things (packages that must be installed, but which may depend on other packages; Terraform resources which may depend on other resources existing; that kind of thing). The output of a topological sort is the vertices of a graph, ordered such that the dependencies of each vertex occur before that vertex.

If you're doing a Linux package update, that means you ensure you install dependencies first; if you're building infrastructure that means Terraform knows to create your VPC before it creates all the stuff that you put inside that VPC.

## TODOs

- smooth out the vertex registration and edge-adding flow -- maybe add a function that takes an adjacency list (or map) and does the uniqueness/presence-checking internally?
    - a lot of these abstractions are being leaked to the package user, which is ugly
- add basic tests
- make this use generics (not just strings)

## Features

- checks to make sure all vertex references are valid when adding edges
- pass in whatever string you want to use for unique vertex Names
- supports multiple dependencies (I guess all implementations do this, so I don't know what I'm celebrating)

## Basic Usage

```go
package main

import (
	"fmt"

	"github.com/groovemonkey/topologicalsort"
)

func main() {
	// Create a new graph
	graph := topologicalsort.NewGraph()

	// Register our packages as vertices
	graph.RegisterVertex("build-essential", "mydata")
	graph.RegisterVertex("gcc", "mydata")
	graph.RegisterVertex("make", "mydata")
	graph.RegisterVertex("libc", "mydata")

	// Add edges to represent dependencies (e.g. build-essential depends on make and gcc)
	graph.AddEdge("build-essential", "make")
	graph.AddEdge("build-essential", "gcc")

	// some of those dependencies have other dependencies (e.g. gcc depends on libc)
	graph.AddEdge("gcc", "libc")

	// perform a topological sort of the graph
	sorted, err := graph.TopologicalSort()
	if err != nil {
		// handle the error
		panic(err)
	}
	fmt.Println("Sorted order:", sorted)
}
```

Running this will get you output like:

```
Sorted order: [make libc gcc build-essential]
```

In other words, if you install the packages in this order, you'll never hit an error due to a missing dependency.


## Advanced Usage: enabling dependencies between HCL config blocks

Let's look at solving a slightly more complicated problem.

Say your program takes an HCL config, which the hclsimple parser brings in as a slice of config objects (in the order they were encountered). That's fine if each piece of config doesn't depend on any other one. However, if you're writing a fancy piece of software and you'd like to let your users write custom config to run (say, by running config-defined commands which may depend on other commands having been run first).

This is kind of a gnarly problem to solve, but it becomes really straightforward when you treat it as a graph/topo sort problem.

Here's a small solution to that problem, built with this graph/toposort library:

```go
package main

import (
	"fmt"
	"log"

	"github.com/groovemonkey/topologicalsort"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type ConfigAction struct {
	Action string `hcl:"action"`
}

type Config struct {
	Name string `hcl:"name"`
	// name of the config block this depends on
	DependsOn []string       `hcl:"dependsOn,optional"`
	Actions   []ConfigAction `hcl:"actions,block"`
}

type HCLConfig struct {
	Configs []Config `hcl:"config,block"`
}

func main() {
	const rawConfig = `
	config {
		name = "becomeroot"
		actions {
			action = "sudo -i"
		}
		dependsOn = ["bootup"]
	}
	config {
		name = "getprocesslist"
		actions {
			action = "ps aux"
		}
		dependsOn = ["becomeroot"]
	}
	config {
		name = "getports"
		actions {
			action = "ss -tulpn"
		}
		dependsOn = ["becomeroot"]
	}
	config {
		name = "checkfile"
		actions {
			action = "ls /tmp/publicfile"
		}
	}
	config {
		name = "checkssh"
		actions {
			action = "if [[ -n $(grep 22 /tmp/portlist) ]]; then cat /etc/ssh/sshd.conf; fi"
		}
		dependsOn = ["becomeroot", "getports"]
	}
	config {
		name = "bootup"
		actions {
			action = "boot the system"
		}
	}
	`

	var parsedConfig HCLConfig
	err := hclsimple.Decode(
		"example.hcl", []byte(rawConfig),
		nil, &parsedConfig,
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}
	fmt.Printf("Configuration is %v\n", parsedConfig)

	// Let's try it!
	sorted, err := topoSortHCL(parsedConfig)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n %v", sorted)
}

func topoSortHCL(parsedConfig HCLConfig) ([]string, error) {
	var returnSlice = []string{}

	// aspirational / API design
	// map[string][]string allows each config (uniquely ID'd by a string) to depend on multiple other configs
	adjacencyList := make(map[string][]string)
	graph := topologicalsort.NewGraph()

	// go through the list of parsed configs once, building up an empty adjacency map
	// this is so we can check for invalid dependencies between configs, which we can't do on the first pass (since we don't yet have a complete set of valid config IDs/names)
	for _, cfg := range parsedConfig.Configs {
		// add the config as a vertex
		// TODO registervertex won't be able to take arbitrary data like RegisterVertex(cfg.Name, cfg) until I add support for generics
		graph.RegisterVertex(cfg.Name, "dummy")

		// add the config ID (name) to our (incomplete) adjacency list
		adjacencyList[cfg.Name] = []string{}
	}

	// iterate through the list of parsed configs again, defining dependencies between our configs and checking for invalid references
	for _, cfg := range parsedConfig.Configs {
		// TODO is this nil? Empty string slice? []string{}?
		if cfg.DependsOn != nil && len(cfg.DependsOn) > 0 {
			for _, d := range cfg.DependsOn {
				// first we check for invalid config references
				_, ok := adjacencyList[cfg.Name]
				if !ok {
					return returnSlice, fmt.Errorf("invalid reference: config %s cannot depend on nonexistent config %s", cfg.Name, d)
				}

				// add the reference
				adjacencyList[cfg.Name] = append(adjacencyList[cfg.Name], d)
			}
		}

	}

	// now that we have built a valid adjacency list, add edges to the graph
	for vertex, deps := range adjacencyList {
		for _, dependency := range deps {
			// NOTE: this builds the directed graph in the traditional way, with a dependency being a vertex with incoming edges
			graph.AddEdge(vertex, dependency)
		}
	}

	// NOTE: cycle detection happens at the beginning of topologicalsort.TopologicalSort
	// sort!
	sortedExecutionOrder, err := graph.TopologicalSort()
	if err != nil {
		return returnSlice, fmt.Errorf("main.go: error during topological sort: %w", err)
	}

	return sortedExecutionOrder, nil
}
```

Running this should result in one of the valid topologically sorted orders, e.g.

```
[bootup becomeroot getprocesslist getports checkfile checkssh]
```

Pretty cool!

