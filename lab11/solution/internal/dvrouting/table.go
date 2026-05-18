package dvrouting

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

type RoutingTable struct {
	Node int
	Dist []int
	Next []int
}

func (rt RoutingTable) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "node %d:\n", rt.Node)
	for d := range rt.Dist {
		if d == rt.Node {
			continue
		}
		if rt.Dist[d] >= Inf {
			fmt.Fprintf(&b, "  -> %d: unreachable\n", d)
			continue
		}
		fmt.Fprintf(&b, "  -> %d: cost=%d via %d\n", d, rt.Dist[d], rt.Next[d])
	}
	return b.String()
}

func FormatTables(tables []RoutingTable) string {
	sorted := slices.Clone(tables)
	slices.SortFunc(sorted, func(a, b RoutingTable) int { return cmp.Compare(a.Node, b.Node) })
	var b strings.Builder
	for _, rt := range sorted {
		b.WriteString(rt.String())
	}
	return b.String()
}
