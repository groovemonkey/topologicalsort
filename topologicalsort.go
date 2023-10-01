package topologicalsort

import (
	"fmt"
)

type Graph[T any] struct {
	// currently a map of graphnode IDs to graphnode pointers
	// could this be map[*GraphNode][]*GraphNode?
	adjacencyList   map[string][]*GraphNode[T]
	vertices        map[string]*GraphNode[T]
	topoSortedOrder []*GraphNode[T]
}

type GraphNode[T any] struct {
	Key  string
	Data T
}

// NewGraph returns an empty graph of the type that's passed in.
func NewGraph[T any](val T) *Graph[T] {
	return &Graph[T]{
		adjacencyList:   make(map[string][]*GraphNode[T]),
		vertices:        make(map[string]*GraphNode[T]),
		topoSortedOrder: make([]*GraphNode[T], 0),
	}
}

func NewGraphNode[T any](key string, data T) *GraphNode[T] {
	return &GraphNode[T]{
		Key:  key,
		Data: data,
	}
}

// RegisterVertex registers a new, unconnected vertex in the graph
func (g *Graph[T]) RegisterVertex(key string, data T) error {
	_, ok := g.vertices[key]
	if ok {
		return fmt.Errorf("attempted to register duplicate vertex")
	}
	// create a new GraphNode and register a pointer to it
	g.vertices[key] = NewGraphNode(key, data)
	return nil
}

// AddEdge adds an edge between two vertices (they need to be looked up by strings, though)
func (g *Graph[T]) AddEdge(source, dest string) error {
	_, ok := g.vertices[source]
	if !ok {
		return fmt.Errorf("attempted to add edge to unregistered vertex %s", source)
	}

	destNode, ok := g.vertices[dest]
	if !ok {
		return fmt.Errorf("attempted to add edge from unregistered vertex %s", dest)
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
func (g *Graph[T]) DepthFirstSearch(node *GraphNode[T], visited, finished map[*GraphNode[T]]bool) (map[*GraphNode[T]]bool, map[*GraphNode[T]]bool, error) {
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
func (g *Graph[T]) SortedKeys() []string {
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
func (g *Graph[T]) SortedValues() []T {
	// create return slice of data from ordered node pointers
	returnSlice := make([]T, len(g.topoSortedOrder))

	// iterate through sorted order and return it
	for i, node := range g.topoSortedOrder {
		returnSlice[i] = (*node).Data
	}
	return returnSlice
}

// TopologicalSort does some basic graph validation (e.g. cycle detection) and then performs a topological sort.
// It returns a slice of strings (the node keys which were originally passed in during graph construction), in a valid topologically sorted order
func (g *Graph[T]) TopologicalSort() ([]string, error) {
	visited := make(map[*GraphNode[T]]bool)
	finished := make(map[*GraphNode[T]]bool)

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

// NewGraphFromData accepts a map of GraphNode:[]string, where the string slice represents adjacent node Keys ("dependencies").
// It returns a graph pointer, or an error if something went wrong.
func NewGraphFromData[T any](nodes map[*GraphNode[T]][]string) (*Graph[T], error) {
	var err error
	graph := &Graph[T]{
		adjacencyList:   make(map[string][]*GraphNode[T]),
		vertices:        make(map[string]*GraphNode[T]),
		topoSortedOrder: make([]*GraphNode[T], 0),
	}
	// Iterate through vertices to build up the graph
	for node := range nodes {
		err = graph.RegisterVertex(node.Key, node.Data)
		if err != nil {
			return nil, err
		}
	}

	// Add edges between vertices
	for node, adjacencies := range nodes {
		for _, a := range adjacencies {
			err = graph.AddEdge(node.Key, a)
			if err != nil {
				return nil, err
			}
		}
	}
	return graph, nil
}

func containsNode[T any](nodes []*GraphNode[T], match *GraphNode[T]) bool {
	for _, n := range nodes {
		if n == match {
			return true
		}
	}
	return false
}
