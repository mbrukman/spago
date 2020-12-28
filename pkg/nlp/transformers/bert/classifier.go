// Copyright 2020 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bert

import (
	"github.com/nlpodyssey/spago/pkg/ml/ag"
	"github.com/nlpodyssey/spago/pkg/ml/nn"
	"github.com/nlpodyssey/spago/pkg/ml/nn/linear"
)

var (
	_ nn.Model = &Classifier{}
)

// ClassifierConfig provides configuration settings for a BERT Classifier.
type ClassifierConfig struct {
	InputSize int
	Labels    []string
}

// Classifier implements a BERT Classifier.
type Classifier struct {
	Config ClassifierConfig
	*linear.Model
}

// NewTokenClassifier returns a new BERT Classifier model.
func NewTokenClassifier(config ClassifierConfig) *Classifier {
	return &Classifier{
		Config: config,
		Model:  linear.New(config.InputSize, len(config.Labels)),
	}
}

// Predict returns the logits.
func (m *Classifier) Predict(xs []ag.Node) []ag.Node {
	return m.Forward(xs...)
}
