package topologicalsort

import (
	"fmt"
)

// TODO make this use generics (not just strings)

type Graph struct {
	// currently a map of graphnode IDs to graphnode pointers
	// could this be map[*GraphNode][]*GraphNode?
	adjacencyList   map[string][]*GraphNode
	vertices        map[string]*GraphNode
	topoSortedOrder []*GraphNode
}

func NewGraph() *Graph {
	return &Graph{
		adjacencyList:   make(map[string][]*GraphNode),
		vertices:        make(map[string]*GraphNode),
		topoSortedOrder: make([]*GraphNode, 0),
	}
}

type GraphNode struct {
	Key  string
	Data string
}

func NewGraphNode(key, data string) *GraphNode {
	return &GraphNode{
		Key:  key,
		Data: data,
	}
}

// RegisterVertex registers a new, unconnected vertex in the graph
func (g *Graph) RegisterVertex(key string, data string) error {
	_, ok := g.vertices[key]
	if ok {
		return fmt.Errorf("attempted to register duplicate vertex")
	}
	// create a new GraphNode and register a pointer to it
	g.vertices[key] = NewGraphNode(key, data)
	return nil
}

// AddEdge adds an edge between two vertices (they need to be looked up by strings, though)
func (g *Graph) AddEdge(source, dest string) error {
	_, ok := g.vertices[source]
	if !ok {
		return fmt.Errorf("attempted to add edge to unregistered vertex %s", source)
	}

	destNode, ok := g.vertices[dest]
	if !ok {
		return fmt.Errorf("attempted to add edge to unregistered vertex %s", source)
	}

	// prevent duplicate additions to adjacencyList
	if containsNode(g.adjacencyList[source], destNode) {
		return fmt.Errorf("attempted to add duplicate edge between %s and %s", source, dest)
	}
	// add edge to adjacencyList
	g.adjacencyList[source] = append(g.adjacencyList[source], destNode)

	return nil
}

// DepthFirstSearch performs a depth-first search starting from vertex node. It uses maps of graphnodes to track which have already been explored and which have been finished
func (g *Graph) DepthFirstSearch(node *GraphNode, visited, finished map[*GraphNode]bool) (map[*GraphNode]bool, map[*GraphNode]bool, error) {
	var err error

	// Mark this node as explored
	visited[node] = true

	for _, neighbor := range g.adjacencyList[node.Key] {
		alreadySeen, ok := visited[neighbor]
		if ok && alreadySeen {
			return nil, nil, fmt.Errorf("\ncycle detected: found a back edge from %s to %s", node.Key, neighbor.Key)
		}

		_, alreadyFinished := finished[neighbor]
		if !alreadyFinished {
			visited, finished, err = g.DepthFirstSearch(neighbor, visited, finished)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	// Remove from visited, mark finished
	visited[node] = false
	finished[node] = true

	g.topoSortedOrder = append(g.topoSortedOrder, node)
	return visited, finished, nil
}

// SortedKeys returns the sorted order of the graph keys
// IT DOES NOT SORT THE GRAPH! (use [TopologicalSort] to do that)
func (g *Graph) SortedKeys() []string {
	// create return slice of keys from ordered node pointers
	returnSlice := make([]string, len(g.topoSortedOrder))

	// iterate through sorted order and return it
	for i, node := range g.topoSortedOrder {
		returnSlice[i] = (*node).Key
	}
	return returnSlice
}

// SortedValues returns the sorted order of the graph values
// IT DOES NOT SORT THE GRAPH! (use [TopologicalSort] to do that)
func (g *Graph) SortedValues() []string {
	// create return slice of data from ordered node pointers
	returnSlice := make([]string, len(g.topoSortedOrder))

	// iterate through sorted order and return it
	for i, node := range g.topoSortedOrder {
		returnSlice[i] = (*node).Data
	}
	return returnSlice
}

// TopologicalSort does some basic graph validation (e.g. cycle detection) and then performs a topological sort.
// It returns a slice of strings (the node keys which were originally passed in during graph construction), in a valid topologically sorted order
func (g *Graph) TopologicalSort() ([]string, error) {
	visited := make(map[*GraphNode]bool)
	finished := make(map[*GraphNode]bool)

	for _, n := range g.vertices {
		_, inVisited := visited[n]
		_, inFinished := finished[n]
		var err error

		// if not yet visited and finished, recurse
		if !inVisited && !inFinished {
			visited, finished, err = g.DepthFirstSearch(n, visited, finished)
			if err != nil {
				return []string{}, err
			}
		}
	}

	// TODO(dcohen) in a future version, just return the topoSortedOrder (pointers, not string Keys or Data)
	return g.SortedKeys(), nil
}

func containsNode(nodes []*GraphNode, match *GraphNode) bool {
	for _, n := range nodes {
		if n == match {
			return true
		}
	}
	return false
}
