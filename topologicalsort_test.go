package topologicalsort

import (
	"fmt"
	"reflect"
	"testing"
)

func graphWithVerticesDUMMYDATA[T any](vertices map[string][]string, sampleData T) *Graph[T] {
	graph := NewGraph(sampleData)

	// Create vertices
	for k := range vertices {
		// uses dummy values
		graph.RegisterVertex(k, sampleData)
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
		dummyData      any
		want           []string
		wantErr        bool
	}{
		{
			name:           "A graph with no vertices is already sorted.",
			adjacency_list: map[string][]string{},
			dummyData:      "",
			want:           []string{},
			wantErr:        false,
		},
		{
			name:           "A graph with one vertex is already sorted.",
			adjacency_list: map[string][]string{"sorted": {}},
			dummyData:      "",
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
			dummyData: "",
			want:      []string{"libc", "gcc", "make", "build-essential"},
			wantErr:   false,
		},
		{
			name: "A graph with no cycles can be sorted.",
			adjacency_list: map[string][]string{
				"four":  {"three"},
				"one":   {},
				"two":   {"one"},
				"three": {"two"},
				"five":  {"four"},
			},
			dummyData: "",
			want:      []string{"one", "two", "three", "four", "five"},
			wantErr:   false,
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
			dummyData: "",
			want:      []string{},
			wantErr:   true,
		},
		{
			name: "Test Generics: a graph using integer data still works",
			adjacency_list: map[string][]string{
				"one": {},
				"two": {"one"},
			},
			dummyData: 0,
			want:      []string{"one", "two"},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := graphWithVerticesDUMMYDATA(tt.adjacency_list, tt.dummyData)
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

func Test_TopographicSort_With_Arbitrary_Data(t *testing.T) {
	type myTestType struct {
		floob string
		burb  int
	}

	tests := []struct {
		name       string
		graph_data map[*GraphNode[myTestType]][]string
		want       []string
		wantErr    bool
	}{
		{
			name:       "A graph with no vertices is already sorted.",
			graph_data: map[*GraphNode[myTestType]][]string{},
			want:       []string{},
			wantErr:    false,
		},
		{
			name: "Graph nodes with struct type Data can still be Topographically sorted",
			graph_data: map[*GraphNode[myTestType]][]string{
				// node three (intentionally out of order)
				{
					Key: "three",
					Data: myTestType{
						floob: "dataforthree",
						burb:  3,
					},
				}: {"two"},
				// node one
				{
					Key: "one",
					Data: myTestType{
						floob: "dataforone",
						burb:  1,
					},
				}: {},
				// node two
				{
					Key: "two",
					Data: myTestType{
						floob: "datafortwo",
						burb:  2,
					},
				}: {"one"},
			},
			want:    []string{"one", "two", "three"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGraphFromData(tt.graph_data)

			got, err := g.TopologicalSort()
			fmt.Println(fmt.Sprintf("%+v", g))
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
