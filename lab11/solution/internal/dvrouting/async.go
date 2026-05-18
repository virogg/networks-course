package dvrouting

import (
	"slices"
	"sync"
	"sync/atomic"
)

const inboxCap = 1024

type vectorMsg struct {
	from int
	dist []int
}

type coordinator struct {
	inflight atomic.Int64
	done     chan struct{}
	once     sync.Once
	nodes    []*asyncNode
}

func (c *coordinator) send(to *asyncNode, m vectorMsg) {
	c.inflight.Add(1)
	to.inbox <- m
}

func (c *coordinator) settled() {
	if c.inflight.Load() != 0 {
		return
	}
	c.once.Do(func() {
		for _, nd := range c.nodes {
			close(nd.inbox)
		}
		close(c.done)
	})
}

type asyncNode struct {
	id    int
	topo  *Topology
	coord *coordinator
	inbox chan vectorMsg
	peers []*asyncNode
	view  map[int][]int
	dist  []int
	next  []int
}

func (nd *asyncNode) recompute() bool {
	n := nd.topo.N
	dist := make([]int, n)
	next := make([]int, n)
	for d := range n {
		if d == nd.id {
			dist[d], next[d] = 0, nd.id
			continue
		}
		best, hop := Inf, -1
		if c := nd.topo.LinkCost(nd.id, d); c < best {
			best, hop = c, d
		}
		for _, peer := range nd.peers {
			vec, ok := nd.view[peer.id]
			if !ok {
				continue
			}
			if c := nd.topo.LinkCost(nd.id, peer.id) + vec[d]; c < best {
				best, hop = c, peer.id
			}
		}
		if best >= Inf {
			best, hop = Inf, -1
		}
		dist[d], next[d] = best, hop
	}
	changed := !slices.Equal(dist, nd.dist)
	nd.dist, nd.next = dist, next
	return changed
}

func (nd *asyncNode) poisonedFor(j int) []int {
	adv := make([]int, len(nd.dist))
	for d := range adv {
		if nd.next[d] == j {
			adv[d] = Inf
		} else {
			adv[d] = nd.dist[d]
		}
	}
	return adv
}

func (nd *asyncNode) broadcast() {
	for _, peer := range nd.peers {
		nd.coord.send(peer, vectorMsg{from: nd.id, dist: nd.poisonedFor(peer.id)})
	}
}

func (nd *asyncNode) loop() {
	for msg := range nd.inbox {
		if !slices.Equal(nd.view[msg.from], msg.dist) {
			nd.view[msg.from] = msg.dist
			if nd.recompute() {
				nd.broadcast()
			}
		}
		nd.coord.inflight.Add(-1)
		nd.coord.settled()
	}
}

type AsyncSimulator struct {
	topo   *Topology
	tables []RoutingTable
}

func NewAsyncSimulator(t *Topology) *AsyncSimulator {
	return &AsyncSimulator{topo: t}
}

func (s *AsyncSimulator) Run() {
	n := s.topo.N
	coord := &coordinator{done: make(chan struct{})}
	nodes := make([]*asyncNode, n)
	for i := range n {
		nodes[i] = &asyncNode{
			id:    i,
			topo:  s.topo,
			coord: coord,
			inbox: make(chan vectorMsg, inboxCap),
			view:  make(map[int][]int),
		}
	}
	for i, nd := range nodes {
		for _, j := range s.topo.Neighbors(i) {
			nd.peers = append(nd.peers, nodes[j])
		}
	}
	coord.nodes = nodes

	coord.inflight.Add(1)
	for _, nd := range nodes {
		nd.recompute()
	}
	for _, nd := range nodes {
		nd.broadcast()
	}

	var wg sync.WaitGroup
	for _, nd := range nodes {
		wg.Go(nd.loop)
	}
	coord.inflight.Add(-1)
	coord.settled()

	<-coord.done
	wg.Wait()

	s.tables = make([]RoutingTable, n)
	for i, nd := range nodes {
		s.tables[i] = RoutingTable{
			Node: i,
			Dist: slices.Clone(nd.dist),
			Next: slices.Clone(nd.next),
		}
	}
}

func (s *AsyncSimulator) UpdateLinkCost(a, b, cost int) {
	s.topo.SetLinkCost(a, b, cost)
	s.Run()
}

func (s *AsyncSimulator) Table(node int) RoutingTable { return s.tables[node] }

func (s *AsyncSimulator) Tables() []RoutingTable { return s.tables }
