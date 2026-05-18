package dvrouting

const Inf = 1 << 16

type Topology struct {
	N    int
	link [][]int
}

func NewTopology(n int) *Topology {
	link := make([][]int, n)
	for i := range link {
		link[i] = make([]int, n)
		for j := range link[i] {
			if i != j {
				link[i][j] = Inf
			}
		}
	}
	return &Topology{N: n, link: link}
}

func (t *Topology) SetLinkCost(a, b, cost int) {
	t.link[a][b] = cost
	t.link[b][a] = cost
}

func (t *Topology) LinkCost(a, b int) int {
	if a == b {
		return 0
	}
	return t.link[a][b]
}

func (t *Topology) Neighbors(i int) []int {
	var ns []int
	for j := range t.N {
		if j != i && t.link[i][j] < Inf {
			ns = append(ns, j)
		}
	}
	return ns
}

func ExampleNetwork() *Topology {
	t := NewTopology(4)
	t.SetLinkCost(0, 1, 1)
	t.SetLinkCost(0, 2, 3)
	t.SetLinkCost(0, 3, 7)
	t.SetLinkCost(1, 2, 1)
	t.SetLinkCost(2, 3, 2)
	return t
}
