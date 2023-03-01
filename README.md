<p align="center">
    <br>
    <img src="https://github.com/nlpodyssey/spago/blob/main/assets/spago_logo.png" width="400"/>
    <br>
<p>
<p align="center">
    <a href="https://github.com/nlpodyssey/spago/actions/workflows/go.yml?query=branch%3Amain">
        <img alt="Build" src="https://github.com/nlpodyssey/spago/actions/workflows/go.yml/badge.svg?branch=main">
    </a>
    <a href="https://codecov.io/gh/nlpodyssey/spago">
        <img alt="Coverage" src="https://codecov.io/gh/nlpodyssey/spago/branch/main/badge.svg">
    </a>
    <a href="https://goreportcard.com/report/github.com/nlpodyssey/spago">
        <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/nlpodyssey/spago">
    </a>
    <a href="https://codeclimate.com/github/nlpodyssey/spago/maintainability">
        <img alt="Maintainability" src="https://api.codeclimate.com/v1/badges/be7350d3eb1a6a8aa503/maintainability">
    </a>
    <a href="https://pkg.go.dev/github.com/nlpodyssey/spago/">
        <img alt="Documentation" src="https://pkg.go.dev/badge/github.com/nlpodyssey/spago/.svg">
    </a>
    <a href="https://opensource.org/licenses/BSD-2-Clause">
        <img alt="License" src="https://img.shields.io/badge/License-BSD%202--Clause-orange.svg">
    </a>
    <a href="http://makeapullrequest.com">
        <img alt="PRs Welcome" src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square">
    </a>
    <a href="https://github.com/avelino/awesome-go">
        <img alt="Awesome Go" src="https://awesome.re/mentioned-badge.svg">
    </a>
</p>
<p align="center">
    <br>
    <i>If you like the project, please ★ star this repository to show your support! 🤩</i>
    <br>
<br>
<p>

> Currently, the main branch contains version v1, which differs substantially from version [v0.7](https://github.com/nlpodyssey/spago/tree/v0.7.0). For NLP-related features, check out the [Cybertron](https://github.com/nlpodyssey/cybertron) package!
> The [CHANGELOG](https://github.com/nlpodyssey/spago/blob/main/CHANGELOG.md) details the major changes.

A **Machine Learning** library written in pure Go designed to support relevant neural architectures in **Natural
Language Processing**.

Spago is self-contained, in that it uses its own lightweight computational graph both for training and
inference, easy to understand from start to finish. 

It provides:
- Automatic differentiation via dynamic define-by-run execution
- Gradient descent optimizers (Adam, RAdam, RMS-Prop, AdaGrad, SGD)
- Feed-forward layers (Linear, Highway, Convolution...)
- Recurrent layers (LSTM, GRU, BiLSTM...)
- Attention layers (Self-Attention, Multi-Head Attention...)
- Memory-efficient Word Embeddings (with [badger](https://github.com/dgraph-io/badger) key–value store)
- Gob compatible neural models for serialization

## Usage

Requirements:

* [Go 1.19](https://golang.org/dl/)

Clone this repo or get the library:

```console
go get -u github.com/nlpodyssey/spago
```

### Dependencies

The core [module](https://github.com/nlpodyssey/spago/blob/v1.0.0-alpha.0/go.mod) of Spago relies only on [testify](github.com/stretchr/testify) for unit testing. 
In other words, it has "zero dependencies", and we are committed to keeping it that way as much as possible.

Spago uses a multi-module [workspace](https://github.com/nlpodyssey/spago/blob/v1.0.0-alpha.0/go.work) to ensure that additional dependencies are downloaded only when specific features (e.g. persistent embeddings) are used.

### Getting Started

A good place to start is by looking at the implementation of built-in neural models, such as the [LSTM](https://github.com/nlpodyssey/spago/blob/af3871a650dddb94299de4f8d3f9eb6ab4fa4a37/nn/recurrent/lstm/lstm.go#L106).
Except for a few linear algebra operations written in assembly for optimal performance (a bit of copying from [Gonum](https://github.com/gonum/gonum)), it's straightforward Go code, so you don't have to worry. In fact, SpaGO could have been written by you :)

The behavior of a neural model is characterized by a combination of parameters and equations. 
Mathematical expressions must be defined using the auto-grad `ag` package in order to take advantage of automatic differentiation.

In this sense, we can say the computational graph is at the center of the Spago machine learning framework.

### Example 1
Here is an example of how to calculate the sum of two variables:

```go
package main

import (
  "fmt"
  "github.com/nlpodyssey/spago/ag"
  "github.com/nlpodyssey/spago/mat"
)

type T = float32

func main() {
  // create a new node of type variable with a scalar
  a := ag.Var(mat.NewScalar(T(2.0))).WithGrad(true)
  // create another node of type variable with a scalar
  b := ag.Var(mat.NewScalar(T(5.0))).WithGrad(true)
  // create an addition operator (the calculation is actually performed here)
  c := ag.Add(a, b)

  // print the result
  fmt.Printf("c = %v (float%d)\n", c.Value(), c.Value().Scalar().BitSize())

  ag.Backward(c, mat.NewScalar(T(0.5)))
  fmt.Printf("ga = %v\n", a.Grad())
  fmt.Printf("gb = %v\n", b.Grad())
}
```

Output:

```console
c = [7] (float32)
ga = [0.5]
gb = [0.5]
```

### Example 2

Here is a simple implementation of the perceptron formula:

```go
package main

import (
  "log"
  "os"

  . "github.com/nlpodyssey/spago/ag"
  "github.com/nlpodyssey/spago/ag/encoding"
  "github.com/nlpodyssey/spago/ag/encoding/dot"
  "github.com/nlpodyssey/spago/mat"
)

func main() {
  x := Var(mat.NewScalar(-0.8)).WithName("x")
  w := Var(mat.NewScalar(0.4)).WithName("w")
  b := Var(mat.NewScalar(-0.2)).WithName("b")

  y := Sigmoid(Add(Mul(w, x), b))

  err := dot.Encode(encoding.NewGraph(y), os.Stdout)
  if err != nil {
    log.Fatal(err)
  }
}

```

In this case, we are interested in rendering the resulting graph with [Graphviz](https://www.graphviz.org/):

```console
go run main.go | dot -Tpng -o g.png
```

<p align="center">
    <br>
    <img src="https://github.com/nlpodyssey/spago/blob/main/assets/perceptron.png" width="400"/>
    <br>
<p>

### Example 3

As a next step, let's take a look at how to create a linear regression model ($y = wx + b$) and how it will be trained.

The following algorithm will try to learn the correct values for weight and bias.  

By the end of our training, our equation will approximate the line of best fit the objective function $y = 3x + 1$.

```go
package main

import (
	"fmt"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/gd"
	"github.com/nlpodyssey/spago/gd/sgd"
	"github.com/nlpodyssey/spago/initializers"
	"github.com/nlpodyssey/spago/losses"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/mat/rand"
	"github.com/nlpodyssey/spago/nn"
)

const (
	epochs   = 100  // number of epochs
	examples = 1000 // number of examples
)

type Linear struct {
	nn.Module
	W nn.Param `spago:"type:weights"`
	B nn.Param `spago:"type:biases"`
}

func NewLinear[T float.DType](in, out int) *Linear {
	return &Linear{
		W: nn.NewParam(mat.NewEmptyDense[T](out, in)),
		B: nn.NewParam(mat.NewEmptyVecDense[T](out)),
	}
}

func (m *Linear) InitWithRandomWeights(seed uint64) *Linear {
	initializers.XavierUniform(m.W.Value(), 1.0, rand.NewLockedRand(seed))
	return m
}

func (m *Linear) Forward(x ag.Node) ag.Node {
	return ag.Add(ag.Mul(m.W, x), m.B)
}

func main() {
	m := NewLinear[float64](1, 1).InitWithRandomWeights(42)
	
	optimizer := gd.NewOptimizer(m, sgd.New[float64](sgd.NewConfig(0.001, 0.9, true)))

	normalize := func(x float64) float64 { return x / float64(examples) }
	objective := func(x float64) float64 { return 3*x + 1 }
	criterion := losses.MSE

	learn := func(input, expected float64) float64 {
		x, target := ag.Scalar(input), ag.Scalar(expected)
		y := m.Forward(x)
		loss := criterion(y, target, true)
		defer ag.Backward(loss) //  free the memory of the graph before return
		return loss.Value().Scalar().F64()
	}

	for epoch := 0; epoch < epochs; epoch++ {
		for i := 0; i < examples; i++ {
			x := normalize(float64(i))
			loss := learn(x, objective(x))
			if i%100 == 0 {
				fmt.Println(loss)
			}
		}
		optimizer.Do()
	}

	fmt.Printf("\nW: %.2f | B: %.2f\n\n", m.W.Value().Scalar().F64(), m.B.Value().Scalar().F64())
}
```

Output:

```console
W: 3.00 | B: 1.00
```

## Performance

Goroutines play a very important role in making Spago efficient; in fact Forward operations are executed concurrently (up to `GOMAXPROCS`). As soon as an Operator is created (usually by calling one of the functions in the `ag` package, such as `Add`, `Prod`, etc.), the related Function's Forward procedure is performed on a new goroutine.
Nevertheless, it's always safe to ask for the Operator's `Value()` without worries: if it's called too soon, the function will lock until the result is computed, and then return the value.

### Known Limits

Sadly, at the moment, Spago is not GPU friendly by design.
 
## Projects using SpaGo

Below is a list of projects that use Spago:

* [Cybertron](https://github.com/nlpodyssey/cybertron) - State-of-the-art Natural Language Processing in Go.
* [Golem](https://github.com/kirasystems/golem) - A batteries-included implementation
  of ["TabNet: Attentive Interpretable Tabular Learning"](https://arxiv.org/abs/1908.07442).
* [Translator](https://github.com/SpecializedGeneralist/translator) - A simple self-hostable Machine Translation
  service.
* [PiSquared](https://github.com/ErikPelli/PiSquared) - A Telegram bot that asks you a question and evaluate the response you provide.
* [WhatsNew](https://github.com/SpecializedGeneralist/whatsnew/) - A simple tool to collect and process quite a few web
  news from multiple sources.

## Contributing

We're glad you're thinking about contributing to Spago! If you think something is missing or could be improved, please
open issues and pull requests. If you'd like to help this project grow, we'd love to have you!

To start contributing, check
the [Contributing Guidelines](https://github.com/nlpodyssey/spago/blob/main/CONTRIBUTING.md).

## Contact

We encourage you to write an issue. This would help the community grow.

If you really want to write to us privately, please email [Matteo Grella](mailto:matteogrella@gmail.com) with your
questions or comments.

## Acknowledgments

Spago is part of the open-source [NLP Odyssey](https://github.com/nlpodyssey) initiative
initiated by members of the EXOP team (now part of Crisis24).

## Sponsors

We appreciate contributions of all kinds. We especially want to thank Spago fiscal sponsors who contribute to ongoing
project maintenance.

See our [Open Collective](https://opencollective.com/nlpodyssey/contribute) page if you too are interested in becoming a sponsor.
