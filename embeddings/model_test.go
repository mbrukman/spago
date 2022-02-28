// Copyright 2022 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package embeddings_test

import (
	"bytes"
	"encoding/gob"
	"github.com/nlpodyssey/spago/embeddings/store"
	"testing"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/embeddings"
	"github.com/nlpodyssey/spago/embeddings/store/memstore"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/nn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ nn.Model[float32] = &embeddings.Model[float32, string]{}

func TestNew(t *testing.T) {
	t.Run("creates a ZeroEmbedding param if UseZeroEmbedding is enabled", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:             5,
			StoreName:        "test-store",
			UseZeroEmbedding: true,
		}
		m := embeddings.New[T, string](conf, repo)

		require.NotNil(t, m.ZeroEmbedding)
		v := m.ZeroEmbedding.Value()
		require.NotNil(t, v)
		assert.Equal(t, 5, v.Rows())
		assert.Equal(t, 1, v.Columns())
		assert.Equal(t, []T{0, 0, 0, 0, 0}, v.Data())
	})

	t.Run("leaves ZeroEmbedding nil if UseZeroEmbedding is disabled", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:             5,
			StoreName:        "test-store",
			UseZeroEmbedding: false,
		}
		m := embeddings.New[T, string](conf, repo)

		require.Nil(t, m.ZeroEmbedding)
	})
}

func TestModel_Count(t *testing.T) {
	type T = float32

	repo := memstore.NewRepository()
	conf := embeddings.Config{
		Size:      1,
		StoreName: "test-store",
		Trainable: true,
	}
	m := embeddings.New[T, string](conf, repo)

	assert.Equal(t, 0, m.Count())

	e, _ := m.Embedding("foo")
	e.ReplaceValue(mat.NewScalar[T](11))
	assert.Equal(t, 1, m.Count())

	e, _ = m.Embedding("bar")
	e.ReplaceValue(mat.NewScalar[T](22))
	assert.Equal(t, 2, m.Count())
}

func TestModel_Embedding(t *testing.T) {
	t.Run("setting a Value causes the embedding to be stored", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:      3,
			StoreName: "test-store",
			Trainable: true,
		}
		m := embeddings.New[T, string](conf, repo)

		e, exists := m.Embedding("e")
		assert.NotNil(t, e)
		assert.False(t, exists)

		// If the embedding is not modified, it should still not exist
		e2, exists := m.Embedding("e")
		assert.NotNil(t, e2)
		assert.False(t, exists)

		// Modify the value
		e.ReplaceValue(mat.NewVecDense([]T{1, 2, 3}))

		// Now it must exist
		e2, exists = m.Embedding("e")
		assert.NotNil(t, e2)
		assert.True(t, exists)
	})

	t.Run("setting a Payload causes the embedding to be stored", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:      1,
			StoreName: "test-store",
			Trainable: true,
		}
		m := embeddings.New[T, string](conf, repo)

		e, exists := m.Embedding("e")
		assert.NotNil(t, e)
		assert.False(t, exists)

		// If the embedding is not modified, it should still not exist
		e2, exists := m.Embedding("e")
		assert.NotNil(t, e2)
		assert.False(t, exists)

		// Modify the value
		e.SetPayload(&nn.Payload[T]{
			Label: 123,
			Data: []mat.Matrix[T]{
				mat.NewScalar[T](11),
				mat.NewScalar[T](22),
			},
		})

		// Now it must exist
		e2, exists = m.Embedding("e")
		assert.NotNil(t, e2)
		assert.True(t, exists)
	})

	t.Run("embeddings with a gradient are memoized", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:      3,
			StoreName: "test-store",
			Trainable: true,
		}
		m := embeddings.New[T, string](conf, repo)

		e, _ := m.Embedding("e")

		e2, _ := m.Embedding("e")
		assert.NotSame(t, e, e2, "no grad: not memoized")

		e.PropagateGrad(mat.NewVecDense([]T{1, 2, 3}))

		e2, _ = m.Embedding("e")
		assert.Same(t, e, e2, "has grad: memoized")

		e.ZeroGrad()

		e2, _ = m.Embedding("e")
		assert.NotSame(t, e, e2, "grad has been zeroed: not memoized")
	})
}

func TestModel_Encode(t *testing.T) {
	t.Run("with UseZeroEmbedding enabled", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:             3,
			StoreName:        "test-store",
			UseZeroEmbedding: true,
		}
		model := embeddings.New[T, string](conf, repo)
		p, g := ag.Reify(model, ag.ForTraining[T]())
		defer g.Clear()

		e, _ := p.Embedding("foo")
		e.ReplaceValue(mat.NewVecDense([]T{1, 2, 3}))

		result := p.Encode([]string{"foo", "bar", "foo"})
		require.Len(t, result, 3)

		assert.NotNil(t, result[0])
		assert.NotNil(t, result[0].Value())
		assert.Equal(t, []T{1, 2, 3}, result[0].Value().Data())

		assert.NotNil(t, result[1])
		assert.NotNil(t, result[1].Value())
		assert.Equal(t, []T{0, 0, 0}, result[1].Value().Data())

		assert.NotNil(t, result[2])
		assert.NotNil(t, result[2].Value())
		assert.Equal(t, []T{1, 2, 3}, result[2].Value().Data())
	})

	t.Run("with UseZeroEmbedding disabled", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:             3,
			StoreName:        "test-store",
			UseZeroEmbedding: false,
		}
		model := embeddings.New[T, string](conf, repo)
		p, g := ag.Reify(model, ag.ForTraining[T]())
		defer g.Clear()

		e, _ := p.Embedding("foo")
		e.ReplaceValue(mat.NewVecDense([]T{1, 2, 3}))

		result := p.Encode([]string{"foo", "bar", "foo"})
		require.Len(t, result, 3)

		assert.NotNil(t, result[0])
		assert.NotNil(t, result[0].Value())
		assert.Equal(t, []T{1, 2, 3}, result[0].Value().Data())

		assert.Nil(t, result[1])

		assert.NotNil(t, result[2])
		assert.NotNil(t, result[2].Value())
		assert.Equal(t, []T{1, 2, 3}, result[2].Value().Data())
	})
}

func TestModel_ClearEmbeddingsWithGrad(t *testing.T) {
	type T = float32

	repo := memstore.NewRepository()
	conf := embeddings.Config{
		Size:      3,
		StoreName: "test-store",
		Trainable: true,
	}
	m := embeddings.New[T, string](conf, repo)

	e, _ := m.Embedding("e")
	e.PropagateGrad(mat.NewVecDense([]T{1, 2, 3}))

	assert.NotNil(t, e.Grad())
	assert.Equal(t, []T{1, 2, 3}, e.Grad().Data())

	m.ClearEmbeddingsWithGrad()

	assert.Nil(t, e.Grad())
}

func TestModel(t *testing.T) {
	t.Run("gob encoding and decoding", func(t *testing.T) {
		type T = float32

		repo := memstore.NewRepository()
		conf := embeddings.Config{
			Size:             3,
			UseZeroEmbedding: true,
			StoreName:        "test-store",
			Trainable:        true,
		}
		m := embeddings.New[T, string](conf, repo)

		e, _ := m.Embedding("e")
		e.ReplaceValue(mat.NewVecDense([]T{1, 2, 3}))
		e.PropagateGrad(mat.NewVecDense([]T{10, 20, 30}))
		e.SetPayload(&nn.Payload[T]{
			Label: 123,
			Data: []mat.Matrix[T]{
				mat.NewScalar[T](11),
				mat.NewScalar[T](22),
			},
		})

		require.NotNil(t, m.ZeroEmbedding)
		require.NotNil(t, m.Store)
		require.Len(t, m.EmbeddingsWithGrad, 1)

		var buf bytes.Buffer
		require.NoError(t, gob.NewEncoder(&buf).Encode(m))

		var decoded *embeddings.Model[T, string]
		require.NoError(t, gob.NewDecoder(&buf).Decode(&decoded))

		require.NotNil(t, decoded)
		assert.Equal(t, conf, decoded.Config)

		require.NotNil(t, decoded.ZeroEmbedding)
		assert.Equal(t, []T{0, 0, 0}, decoded.ZeroEmbedding.Value().Data())

		require.NotNil(t, decoded.Store)
		assert.Nil(t, decoded.Store.(store.PreventStoreMarshaling).Store)

		require.Nil(t, decoded.EmbeddingsWithGrad)
	})
}
