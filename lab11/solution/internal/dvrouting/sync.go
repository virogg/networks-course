package dvrouting

import "slices"

type SyncSimulator struct {
	topo *Topology
	dist [][]int
	next [][]int
}

func NewSyncSimulator(t *Topology) *SyncSimulator {
	return &SyncSimulator{topo: t}
}

func (s *SyncSimulator) Run() int {
	n := s.topo.N
	s.dist = make([][]int, n)
	s.next = make([][]int, n)
	for i := range n {
		s.dist[i], s.next[i] = s.initVector(i)
	}

	rounds := 0
	for {
		rounds++
		changed := false
		newDist := make([][]int, n)
		newNext := make([][]int, n)
		for i := range n {
			d, nh := s.relax(i)
			newDist[i], newNext[i] = d, nh
			if !slices.Equal(d, s.dist[i]) {
				changed = true
			}
		}
		s.dist, s.next = newDist, newNext
		if !changed {
			return rounds
		}
	}
}

func (s *SyncSimulator) initVector(i int) (dist, next []int) {
	n := s.topo.N
	dist = make([]int, n)
	next = make([]int, n)
	for d := range n {
		switch c := s.topo.LinkCost(i, d); {
		case d == i:
			dist[d], next[d] = 0, i
		case c < Inf:
			dist[d], next[d] = c, d
		default:
			dist[d], next[d] = Inf, -1
		}
	}
	return dist, next
}

func (s *SyncSimulator) relax(i int) (dist, next []int) {
	n := s.topo.N
	dist = make([]int, n)
	next = make([]int, n)
	neighbors := s.topo.Neighbors(i)
	for d := range n {
		if d == i {
			dist[d], next[d] = 0, i
			continue
		}
		best, hop := Inf, -1
		for _, j := range neighbors {
			if c := s.topo.LinkCost(i, j) + s.dist[j][d]; c < best {
				best, hop = c, j
			}
		}
		dist[d], next[d] = min(best, Inf), hop
	}
	return dist, next
}

func (s *SyncSimulator) UpdateLinkCost(a, b, cost int) int {
	s.topo.SetLinkCost(a, b, cost)
	return s.Run()
}

func (s *SyncSimulator) Table(node int) RoutingTable {
	return RoutingTable{
		Node: node,
		Dist: slices.Clone(s.dist[node]),
		Next: slices.Clone(s.next[node]),
	}
}

func (s *SyncSimulator) Tables() []RoutingTable {
	out := make([]RoutingTable, s.topo.N)
	for i := range out {
		out[i] = s.Table(i)
	}
	return out
}
