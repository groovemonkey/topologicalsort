package topologicalsort

import (
	"reflect"
	"testing"
)

func graphWithVertices(vertices map[string][]string) *Graph {
	graph := NewGraph()

	// Create vertices
	for k := range vertices {
		// uses dummy values
		graph.RegisterVertex(k, "")
	}

	// add edges
	for k, v := range vertices {
		for _, vert := range v {
			// fmt.Println(fmt.Sprintf("%s depends on %s", k, vert))
			graph.AddEdge(k, vert)
		}

	}
	return graph
}

func TestGraph_TopologicalSort(t *testing.T) {
	tests := []struct {
		name string
		// adjacency_list is used to build the test graph -- vertices and all of their dependencies
		adjacency_list map[string][]string
		want           []string
		wantErr        bool
	}{
		{
			name:           "A graph with no vertices is already sorted.",
			adjacency_list: map[string][]string{},
			want:           []string{},
			wantErr:        false,
		},
		{
			name:           "A graph with one vertex is already sorted.",
			adjacency_list: map[string][]string{"sorted": {}},
			want:           []string{"sorted"},
			wantErr:        false,
		},
		// // two valid permutations of sorted order, so skipping
		// {
		// 	name:     "A graph with two unconnected vertices is always sorted.",
		// 	vertices: map[string][]string{"sorted": {}, "fine": {}},
		// 	want:     []string{"sorted", "fine"},
		// 	wantErr:  false,
		// },
		{
			name: "Package manager example from cmd",
			adjacency_list: map[string][]string{
				"build-essential": {"make", "gcc"},
				"make":            {"gcc"},
				"gcc":             {"libc"},
				"libc":            {},
			},
			want:    []string{"libc", "gcc", "make", "build-essential"},
			wantErr: false,
		},
		{
			name: "A graph with no cycles can be sorted.",
			adjacency_list: map[string][]string{
				"one":   {},
				"two":   {"one"},
				"three": {"two"},
				"four":  {"three"},
				"five":  {"four"},
			},
			want:    []string{"one", "two", "three", "four", "five"},
			wantErr: false,
		},
		{
			name: "A graph with a cycle triggers an error",
			adjacency_list: map[string][]string{
				"one":   {},
				"cycle": {"one", "three"},
				"three": {"cycle", "one"},
				"four":  {"three", "two", "one"},
				"five":  {"four", "three"},
			},
			want:    []string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := graphWithVertices(tt.adjacency_list)
			got, err := g.TopologicalSort()
			if (err != nil) != tt.wantErr {
				t.Errorf("Graph.TopologicalSort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Graph.TopologicalSort() = %v, want %v", got, tt.want)
			}
		})
	}
}
