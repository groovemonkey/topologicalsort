package topologicalsort

import (
	"fmt"
)

// TODO make this use generics (not just strings)

type Graph struct {
	// TODO length could help make sure all nodes are connected?

	// currently a map of graphnode IDs to graphnode pointers
	// would this be map[*GraphNode][]*GraphNode?
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
	// TODO should we just allow this and then somehow check the graph to make sure it's connected? (possibly using 'len'?)
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

// TODO implement cycle detection
func (g *Graph) IsCyclic() bool {
	return false
}

// DepthFirstSearch performs a depth-first search starting from vertex startV. It uses a map of graphnodes to track which have already been explored
func (g *Graph) DepthFirstSearch(startV *GraphNode, exploredMap *map[*GraphNode]bool) {
	explored := *exploredMap

	// Mark this node as explored
	explored[startV] = true

	for _, node := range startV.outgoingEdges {
		if !explored[node] {
			g.DepthFirstSearch(node, &explored)
		}
	}
	g.topoSortedOrder = append(g.topoSortedOrder, startV)
}

// TopologicalSort does some basic graph validation (e.g. cycle detection) and then performs a topological sort.
// It returns a slice of strings (the node keys which were originally passed in during graph construction), in a valid topologically sorted order
func (g *Graph) TopologicalSort() ([]string, error) {
	numVertices := len(g.vertices)
	returnSlice := make([]string, numVertices)

	// TODO: do we at least have one node without incoming edges?
	if g.IsCyclic() {
		return returnSlice, fmt.Errorf("attempted topological sort of a cyclic graph. Topological sort only works on directed, acyclic graphs (DAGs).")
	}

	// exploredNodes tracks which nodes we've visited and finished exploring
	exploredNodes := make(map[*GraphNode]bool, numVertices)

	for _, v := range g.vertices {
		// if not yet explored
		_, ok := exploredNodes[v]
		if !ok {
			numVertices--
			g.DepthFirstSearch(v, &exploredNodes)
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
