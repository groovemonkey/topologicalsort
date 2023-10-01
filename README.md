# Graph / Topological Sort

This is a small library I built to scratch a personal itch.

## What is topological sorting?

A directed acyclic graph (DAG) is often used to represent dependencies between things (packages that must be installed, but which may depend on other packages; Terraform resources which may depend on other resources existing; that kind of thing). The output of a topological sort is the vertices of a graph, ordered such that the dependencies of each vertex occur before that vertex.

If you're installing a Linux package, this means you have to ensure that dependencies (and all their dependencies) are installed before you install the package; if you're building infrastructure it means Terraform must create your VPC before it creates all the stuff which gets placed inside that VPC. Topological sort is an algorithm that quickly works out a valid order.

## Features

- uses generics for Node Data (attach any type of data you like!)
- checks to make sure all vertex references are valid when adding edges
- uses strings for vertex keys
- supports multiple dependencies (I guess all implementations do this, so I don't know what I'm celebrating, but this was the original itch I wanted to scratch)

## Primitives
- `NewGraph(T type)` creates an empty graph where items contain data of type `type`
- add items (vertices) with `AddItem` or `RegisterVertex` (they are equivalent)
- add dependencies (edges) with `AddDependency` or `AddEdge` (they are equivalent)
- `NewGraphFromData` is a constructor which creates a graph from structured input data.

## Basic Usage

```go
package main

import (
	"fmt"

	"github.com/groovemonkey/topologicalsort"
)

func main() {
	// Create a new graph, passing in the type of data we want to use for our nodes
	graph := topologicalsort.NewGraph("")

	// Register our packages as vertices
	graph.AddItem("build-essential", "be-data")
	graph.AddItem("gcc", "gcc-data")
	// AddItem and RegisterVertex are equivalent
	graph.RegisterVertex("make", "make-data")
	graph.RegisterVertex("libc", "libc-data")

	// Add edges to represent dependencies (e.g. build-essential depends on make and gcc)
	graph.AddDependency("build-essential", "make")
	graph.AddDependency("build-essential", "gcc")

	// some of those dependencies have other dependencies (e.g. gcc depends on libc)
	// AddDependency and AddEdge are equivalent
	graph.AddEdge("gcc", "libc")

	// perform a topological sort of the graph
	sorted, err := graph.TopologicalSort()
	if err != nil {
		// handle the error
		panic(err)
	}
	fmt.Println("Sorted Keys:", sorted, "are also available via", graph.SortedKeys())
	fmt.Println("Sorted Data:", graph.SortedValues())
}
```

Running this will get you output like:

```
Sorted Keys: [libc make gcc build-essential] are also available via [libc make gcc build-essential]
Sorted Data: [libc-data make-data gcc-data be-data]
```

In practical terms, if you install the packages in this order, you'll never hit an error due to a missing dependency.


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

	// Create a new Graph[string] by passing a string to the constructor
	graph := topologicalsort.NewGraph("")

	// go through the list of parsed configs once, building up an empty adjacency map
	// this is so we can check for invalid dependencies between configs, which we can't do on the first pass (since we don't yet have a complete set of valid config IDs/names)
	for _, cfg := range parsedConfig.Configs {
		// add the config as a vertex
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


## TODOs

- TODO(dcohen) in a future version, `TopologicalSort()` should return the `graph.topoSortedOrder` (pointers, not string Keys or Data)
- should the graph even keep a toposorted order? Or should that be dynamically generated and immediately returned?
- does the graph struct make sense? Do we need vertices AND an adjacencylist? I think just the adjacencylist gives us vertices (adjacencylist keys are vertices)

