// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ag

import (
	"fmt"
	"github.com/nlpodyssey/spago/ag/fn"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/rand"
	"github.com/nlpodyssey/spago/utils/processingqueue"
	"sync"
)

// The Graph a.k.a. expression graph or computational graph is the centerpiece of the spaGO machine learning framework.
// It takes the form of a directed graph with no directed cycles (DAG).
type Graph[T mat.DType] struct {
	// to avoid data race during concurrent computations (mu2 is used in Constant())
	mu, mu2 sync.Mutex
	// maxID is the id of the last inserted node (corresponds of len(nodes)-1)
	maxID int
	// the time-step is useful to perform truncated back propagation (default 0)
	curTimeStep int
	// nodes contains the list of nodes of the graph. The indices of the list are the nodes ids.
	// The nodes are inserted one at a time in order of creation.
	nodes []Node[T]
	// constants maps scalar values that that doesn't require gradients to a Node. It is used in the Constant() method.
	constants map[T]Node[T]
	// eagerExecution reports whether to compute the forward during the graph definition.
	eagerExecution bool
	// cache of the support structures created during the last groupNodesByHeight() computation.
	// Before using it you have to check if the maxID of the graph matches the maxID of the cache.
	// Otherwise, the cache must be invalidated and the values recalculated.
	cache struct {
		// the maxID when this cache was created.
		maxID int
		// nodes grouped by height
		nodesByHeight [][]Node[T]
		// the nodes height. The index corresponds to the node ID.
		height []int
	}
	// randGen is the generator of random numbers
	randGen *rand.LockedRand[T]
	// processingQueue allows proper handling for computationally heavy operations
	// such as forward and backward steps.
	// The default size is defaultProcessingQueueSize.
	processingQueue processingqueue.ProcessingQueue
}

// NewGraph returns a new initialized graph.
// It can take an optional random generator of type rand.WithRand.
func NewGraph[T mat.DType](opts ...GraphOption[T]) *Graph[T] {
	g := &Graph[T]{
		maxID:           -1,
		curTimeStep:     0,
		nodes:           nil,
		constants:       map[T]Node[T]{},
		eagerExecution:  true,
		processingQueue: processingqueue.New(defaultProcessingQueueSize),
	}
	g.clearCache()
	for _, opt := range opts {
		opt(g)
	}
	if g.randGen == nil {
		g.randGen = rand.NewLockedRand[T](1) // set default random generator
	}
	return g
}

// SetRand replace the graph's random number generator with the given one.
func (g *Graph[T]) SetRand(rand *rand.LockedRand[T]) {
	g.randGen = rand
}

// SetRandSeed replace the graph's random number generator with a new one with the given seed.
func (g *Graph[T]) SetRandSeed(seed uint64) {
	g.randGen = rand.NewLockedRand[T](seed)
}

// SetEagerExecution sets whether to compute the forward during the graph definition (default true).
// When enabled it lets you access to the Value() resulting from the computation.
// There are particular cases where you don't need intermediate values so computing the forward after
// the graph definition can be more efficient.
// It returns the previous value.
func (g *Graph[T]) SetEagerExecution(value bool) bool {
	prev := g.eagerExecution
	g.eagerExecution = value
	return prev
}

// SetConcurrentComputations sets the maximum number of concurrent computations handled by the Graph
// for heavy tasks such as forward and backward steps.
// The value 1 corresponds to sequential execution.
// It returns the previous value.
func (g *Graph[T]) SetConcurrentComputations(value int) int {
	if value < 1 {
		panic("ag: WithConcurrentComputations value must be greater than zero")
	}
	prev := len(g.processingQueue)
	g.processingQueue = processingqueue.New(value)
	return prev
}

// EagerExecutionEnabled returns whether the computation happens during the graph definition.
// See ag.WithEagerExecution() option.
func (g *Graph[_]) EagerExecutionEnabled() bool {
	return g.eagerExecution
}

// ConcurrentComputations returns the maximum number of concurrent computations handled by the Graph
// for heavy tasks such as forward and backward steps.
func (g *Graph[_]) ConcurrentComputations() int {
	return g.processingQueue.Size()
}

// ZeroGrad sets the gradients of all nodes to zero.
func (g *Graph[_]) ZeroGrad() {
	for _, node := range g.nodes {
		node.ZeroGrad()
	}
}

// IncTimeStep increments the value of the graph's TimeStep by one.
func (g *Graph[_]) IncTimeStep() {
	g.curTimeStep++
}

// TimeStep is an integer value associated with the graph, which can be useful
// to perform truncated back propagation. This value is 0 for a new Graph, and
// can be incremented calling IncTimeStep.
func (g *Graph[_]) TimeStep() int {
	return g.curTimeStep
}

// Nodes returns the nodes of the graph.
func (g *Graph[T]) Nodes() []Node[T] {
	return g.nodes
}

// Clear cleans the graph. This is a destructive operation.
// It is not mandatory to call this method, but it is strongly recommended to do so when you finish using the graph.
// The cleaning of the graph improves the memory management and therefore the efficiency of execution.
// Clear releases the matrices underlying the nodes so to reduce the need of future new time-consuming allocations.
// It is important to stress that calling g.Clean(), the "value" and "grad" of the operators nodes are freed (set to nil).
// Whoever is using the Value() or Grad() properties of a node, does so at his own risk. It is therefore recommended to
// make always a copy of the return value of Value() or Grad().
// Alternatively, you can use the convenient graph's methods g.CopyValue(node) and g.CopyGrad(node).
func (g *Graph[T]) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.nodes == nil {
		return
	}
	g.maxID = -1
	g.curTimeStep = 0
	g.clearCache()
	g.releaseMemory()

	for _, node := range g.nodes {
		if node, ok := node.(*Operator[T]); ok {
			*node = Operator[T]{}
			getOperatorPool[T]().Put(node)
		}
	}

	g.nodes = nil
}

// ClearForReuse does the same thing as Clear(), with the difference that the graph structure (i.e.
// how nodes are connected to each other) is maintained.
// This allows you to efficiently use the graph as if it were "pre-computed" (see the Forward()
// method for this scenario).
func (g *Graph[_]) ClearForReuse() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.nodes == nil {
		return
	}
	g.releaseMemory()
}

// NewVariable creates and returns a new node.
func (g *Graph[T]) NewVariable(value mat.Matrix[T], requiresGrad bool) Node[T] {
	n := &Variable[T]{
		graph:        g,
		timeStep:     g.curTimeStep,
		value:        value,
		grad:         nil,
		requiresGrad: requiresGrad,
	}
	return g.insert(n)
}

// NewVariableWithName creates and returns a new node.
func (g *Graph[T]) NewVariableWithName(value mat.Matrix[T], requiresGrad bool, name string) Node[T] {
	n := &Variable[T]{
		graph:        g,
		timeStep:     g.curTimeStep,
		name:         name,
		value:        value,
		grad:         nil,
		requiresGrad: requiresGrad,
	}
	return g.insert(n)
}

// NewScalar creates a variable node that doesn't require gradients.
// TODO: Why shouldn't gradient be required by default?
func (g *Graph[T]) NewScalar(value T) Node[T] {
	return g.NewVariable(mat.NewScalar(value), false)
}

// NewScalarWithName creates a variable node that doesn't require gradients.
// TODO: Why shouldn't gradient be required by default?
func (g *Graph[T]) NewScalarWithName(value T, name string) Node[T] {
	return g.NewVariableWithName(mat.NewScalar(value), false, name)
}

// Constant returns a scalar Node that that doesn't require gradients.
// For the same value, a previously created Node is returned without creating a new one.
// Useful for example in the case of epsilon and number like 0.0 or 1.0.
func (g *Graph[T]) Constant(value T) Node[T] {
	g.mu2.Lock()
	defer g.mu2.Unlock()
	if node, ok := g.constants[value]; ok {
		return node
	}
	node := g.NewVariableWithName(mat.NewScalar(value), false, fmt.Sprint(value))
	g.constants[value] = node
	return node
}

// NewOperator creates a new operator along with its forward pass.
// Please note that operations must be performed among nodes belonging to the same graph; it panics otherwise.
func (g *Graph[T]) NewOperator(f fn.Function[T], operands ...Node[T]) Node[T] {
	for _, o := range operands {
		if o.Graph() != g {
			panic("ag: operations cannot be executed among nodes of different graphs. " +
				"You may consider wrapping the nodes you need with NewWrap().")
		}
	}
	var value mat.Matrix[T] = nil
	if g.eagerExecution {
		// the calculation is out of the lock, so it can run concurrently with other operators
		g.processingQueue.Run(func() {
			value = f.Forward()
		})
	}
	requiresGrad := false
	for _, operand := range operands {
		if operand.RequiresGrad() {
			requiresGrad = true
			break
		}
	}

	n := getOperatorPool[T]().Get().(*Operator[T])
	*n = Operator[T]{
		graph:        g,
		timeStep:     g.curTimeStep,
		function:     f,
		operands:     operands,
		value:        value,
		grad:         nil,
		requiresGrad: requiresGrad,
	}
	return g.insert(n)
}

// NewWrap creates a new wrapper Node for the given value, attaching it to
// the graph.
func (g *Graph[T]) NewWrap(value GradValue[T]) Node[T] {
	n := &Wrapper[T]{
		GradValue: value,
		timeStep:  g.curTimeStep,
		graph:     g,
		wrapGrad:  true,
	}
	return g.insert(n)
}

// NewWrapNoGrad is similar to NewWrap, but it disables automatic
// differentiation on the new node.
func (g *Graph[T]) NewWrapNoGrad(value GradValue[T]) Node[T] {
	n := &Wrapper[T]{
		GradValue: value,
		graph:     g,
		timeStep:  g.curTimeStep,
		wrapGrad:  false,
	}
	return g.insert(n)
}

// nodeInternal extends the public Node with private methods.
type nodeInternal[T mat.DType] interface {
	Node[T]
	setID(int)
}

// insert append the node into the graph's nodes and assign it an id.
func (g *Graph[T]) insert(n nodeInternal[T]) Node[T] {
	g.mu.Lock()
	g.maxID++
	n.setID(g.maxID)
	g.nodes = append(g.nodes, n)
	g.mu.Unlock()
	return n
}

// clearCache cleans the cache.
func (g *Graph[_]) clearCache() {
	g.cache.maxID = -1
	g.cache.nodesByHeight = nil
	g.cache.height = nil
}

// releaseMemory clears the values and the gradients of operator nodes.
// Since the values and the gradients within the nodes are handled through a pool of dense matrices,
// releasing them allows the memory to be reused without being reallocated, improving performance.
func (g *Graph[T]) releaseMemory() {
	for _, node := range g.nodes {
		if node, ok := node.(*Operator[T]); ok {
			g.releaseValue(node)
			g.releaseGrad(node)
		}
	}
}

// releaseValue set the node value to nil release the memory.
func (g *Graph[T]) releaseValue(node *Operator[T]) {
	if node.value == nil {
		return
	}
	mat.ReleaseMatrix(node.value)
	node.value = nil
}

// releaseGrad set the node gradient to nil and release the memory.
func (g *Graph[T]) releaseGrad(node *Operator[T]) {
	node.ZeroGrad()
}

func (g *Graph[T]) groupNodesByHeight() [][]Node[T] {
	if g.cache.maxID == g.maxID {
		return g.cache.nodesByHeight
	}
	groups := g.cache.nodesByHeight
	height := make([]int, len(g.nodes))
	copy(height[:len(g.cache.height)], g.cache.height)

	startIndex := g.cache.maxID + 1
	for _, node := range g.nodes[startIndex:] {
		h := 0
		if node, ok := node.(*Operator[T]); ok {
			for _, operand := range node.operands {
				if operand, ok := operand.(*Operator[T]); ok {
					if height[operand.id] >= h {
						h = height[operand.id] + 1
					}
				}
			}
		}
		height[node.ID()] = h
		if h == len(groups) {
			groups = append(groups, make([]Node[T], 0, 1))
		}
		groups[h] = append(groups[h], node)
	}

	// update cache and return
	g.cache.maxID = g.maxID
	g.cache.nodesByHeight = groups
	g.cache.height = height
	return groups
}

// MarshalBinary satisfies encoding.BinaryMarshaler interface and prevents
// a Graph to be encoded to binary representation.
// This is relevant in the context of a Graph being part of a nn.Model: when
// serializing a model to binary, we want to skip the Graph, since it is part
// of the runtime context only.
func (g *Graph[_]) MarshalBinary() ([]byte, error) {
	return []byte{}, nil
}

// UnmarshalBinary satisfies encoding.BinaryUnmarshaler interface.
func (g *Graph[_]) UnmarshalBinary(_ []byte) error {
	return nil
}
