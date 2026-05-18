package dvrouting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var initialDist = [][]int{
	{0, 1, 2, 4},
	{1, 0, 1, 3},
	{2, 1, 0, 2},
	{4, 3, 2, 0},
}
var initialNext = [][]int{
	{0, 1, 1, 1},
	{0, 1, 2, 2},
	{1, 1, 2, 3},
	{2, 2, 2, 3},
}

var updatedDist = [][]int{
	{0, 1, 3, 5},
	{1, 0, 4, 6},
	{3, 4, 0, 2},
	{5, 6, 2, 0},
}
var updatedNext = [][]int{
	{0, 1, 2, 2},
	{0, 1, 0, 0},
	{0, 0, 2, 3},
	{2, 2, 2, 3},
}

func assertTables(t *testing.T, tables []RoutingTable, dist, next [][]int) {
	t.Helper()
	require.Len(t, tables, len(dist))
	for _, rt := range tables {
		require.Equal(t, dist[rt.Node], rt.Dist, "node %d distances", rt.Node)
		require.Equal(t, next[rt.Node], rt.Next, "node %d next hops", rt.Node)
	}
}

// Task А
func TestSyncConverges(t *testing.T) {
	sim := NewSyncSimulator(ExampleNetwork())
	require.Positive(t, sim.Run())
	assertTables(t, sim.Tables(), initialDist, initialNext)
}

// Task Б
func TestSyncRecomputesOnLinkChange(t *testing.T) {
	sim := NewSyncSimulator(ExampleNetwork())
	sim.Run()
	assertTables(t, sim.Tables(), initialDist, initialNext)

	require.Positive(t, sim.UpdateLinkCost(1, 2, 8))
	assertTables(t, sim.Tables(), updatedDist, updatedNext)
}

// Task В
func TestAsyncConverges(t *testing.T) {
	sim := NewAsyncSimulator(ExampleNetwork())
	sim.Run()
	assertTables(t, sim.Tables(), initialDist, initialNext)
}

// Task В + Б
func TestAsyncRecomputesOnLinkChange(t *testing.T) {
	sim := NewAsyncSimulator(ExampleNetwork())
	sim.Run()
	assertTables(t, sim.Tables(), initialDist, initialNext)

	sim.UpdateLinkCost(1, 2, 8)
	assertTables(t, sim.Tables(), updatedDist, updatedNext)
}

func TestSyncAndAsyncAgree(t *testing.T) {
	for range 50 {
		sync := NewSyncSimulator(ExampleNetwork())
		sync.Run()
		async := NewAsyncSimulator(ExampleNetwork())
		async.Run()
		for n := range 4 {
			require.Equal(t, sync.Table(n).Dist, async.Table(n).Dist, "node %d", n)
			require.Equal(t, sync.Table(n).Next, async.Table(n).Next, "node %d", n)
		}
	}
}

func TestLinkCostDecreaseShortensRoute(t *testing.T) {
	sim := NewSyncSimulator(ExampleNetwork())
	sim.Run()
	require.Equal(t, 4, sim.Table(0).Dist[3])

	sim.UpdateLinkCost(0, 3, 2)
	require.Equal(t, 2, sim.Table(0).Dist[3])
	require.Equal(t, 3, sim.Table(0).Next[3])
}
