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
	Name          string
	Data          string
	incomingEdges []*GraphNode
	outgoingEdges []*GraphNode
}

func NewGraphNode(name, data string) *GraphNode {
	return &GraphNode{
		Name: name,
		Data: data,
		// incoming and outgoing edges are empty, will be added via [AddEdge]
		incomingEdges: make([]*GraphNode, 0),
		outgoingEdges: make([]*GraphNode, 0),
	}
}

// RegisterVertex registers a new, unconnected vertex in the graph
func (g *Graph) RegisterVertex(name string, data string) error {
	_, ok := g.vertices[name]
	if ok {
		return fmt.Errorf("attempted to register duplicate vertex")
	}
	// create a new GraphNode and register a pointer to it
	g.vertices[name] = NewGraphNode(name, data)
	return nil
}

// AddEdge adds an edge between two vertices (they need to be looked up by strings, though)
func (g *Graph) AddEdge(source, dest string) error {
	// TODO should we just autoregister by calling RegisterVertex from here?
	sourceNode, ok := g.vertices[source]
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

	// update node edges
	sourceNode.outgoingEdges = append(sourceNode.outgoingEdges, destNode)
	destNode.incomingEdges = append(destNode.incomingEdges, sourceNode)
	return nil
}

// DepthFirstSearch performs a depth-first search starting from vertex node. It uses maps of graphnodes to track which have already been explored and which have been finished
func (g *Graph) DepthFirstSearch(node *GraphNode, visited, finished map[*GraphNode]bool) (map[*GraphNode]bool, map[*GraphNode]bool, error) {
	var err error

	// Mark this node as explored
	visited[node] = true

	for _, neighbor := range g.adjacencyList[node.Name] {
		_, alreadySeen := visited[neighbor]
		if alreadySeen {
			return nil, nil, fmt.Errorf("\ncycle detected: found a back edge from %s to %s", node.Name, neighbor.Name)
		}

		_, alreadyFinished := finished[neighbor]
		if !alreadyFinished {
			visited, finished, err = g.DepthFirstSearch(neighbor, visited, finished)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	// visited[node] = false
	finished[node] = true

	g.topoSortedOrder = append(g.topoSortedOrder, node)
	return visited, finished, nil
}

// TopologicalSort does some basic graph validation (e.g. cycle detection) and then performs a topological sort.
// It returns a slice of strings (the node keys which were originally passed in during graph construction), in a valid topologically sorted order
func (g *Graph) TopologicalSort() ([]string, error) {
	numVertices := len(g.vertices)
	returnSlice := make([]string, numVertices)
	visited := make(map[*GraphNode]bool)
	finished := make(map[*GraphNode]bool)

	for v, _ := range g.vertices {
		n := g.vertices[v]
		// if not yet visited
		_, inVisited := visited[n]
		_, inFinished := finished[n]
		var err error

		if !inVisited && !inFinished {
			visited, finished, err = g.DepthFirstSearch(n, visited, finished)
			if err != nil {
				return returnSlice, err
			}
		}
	}

	// create return slice of names from ordered node pointers
	for i, n := range g.topoSortedOrder {
		node := *n
		returnSlice[i] = node.Name
	}
	return returnSlice, nil
}

func containsNode(nodes []*GraphNode, match *GraphNode) bool {
	for _, n := range nodes {
		if n == match {
			return true
		}
	}
	return false
}
