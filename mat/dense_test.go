// Copyright 2022 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ Matrix[float32] = &Dense[float32]{}
var _ Matrix[float64] = &Dense[float64]{}

func TestNewDense(t *testing.T) {
	t.Run("float32", testNewDense[float32])
	t.Run("float64", testNewDense[float64])
}

func testNewDense[T DType](t *testing.T) {
	t.Run("negative rows", func(t *testing.T) {
		require.Panics(t, func() {
			NewDense(-1, 1, []T{})
		})
	})

	t.Run("negative cols", func(t *testing.T) {
		require.Panics(t, func() {
			NewDense(1, -1, []T{})
		})
	})

	t.Run("elements length mismatch", func(t *testing.T) {
		require.Panics(t, func() {
			NewDense(1, 1, []T{1, 2})
		})
	})

	testCases := []struct {
		r int
		c int
		e []T
	}{
		{0, 0, nil},
		{0, 0, []T{}},

		{0, 1, nil},
		{0, 1, []T{}},

		{1, 0, nil},
		{1, 0, []T{}},

		{1, 1, []T{1}},
		{1, 2, []T{1, 2}},
		{2, 1, []T{1, 2}},
		{2, 2, []T{1, 2, 3, 4}},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d, %d, %#v", tc.r, tc.c, tc.e), func(t *testing.T) {
			d := NewDense(tc.r, tc.c, tc.e)
			assertDenseDims(t, tc.r, tc.c, d)
			assert.Len(t, d.Data(), len(tc.e))
			if tc.e != nil {
				assert.Equal(t, tc.e, d.Data())
			}
		})
	}

	t.Run("data is copied", func(t *testing.T) {
		s := []T{1}
		d := NewDense(1, 1, s)
		s[0] = 42 // modifying s must not modify d.data
		assert.Equal(t, T(1), d.data[0])
	})
}

func TestNewVecDense(t *testing.T) {
	t.Run("float32", testNewVecDense[float32])
	t.Run("float64", testNewVecDense[float64])
}

func testNewVecDense[T DType](t *testing.T) {
	testCases := [][]T{
		nil,
		{},
		{1},
		{1, 2},
		{1, 2, 3},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%#v", tc), func(t *testing.T) {
			d := NewVecDense(tc)
			assertDenseDims(t, len(tc), 1, d)
			assert.Len(t, d.Data(), len(tc))
			if tc != nil {
				assert.Equal(t, tc, d.Data())
			}
		})
	}

	t.Run("data is copied", func(t *testing.T) {
		s := []T{1}
		d := NewVecDense(s)
		s[0] = 42 // modifying s must not modify d.data
		assert.Equal(t, T(1), d.data[0])
	})
}

func TestNewScalar(t *testing.T) {
	t.Run("float32", testNewScalar[float32])
	t.Run("float64", testNewScalar[float64])
}

func testNewScalar[T DType](t *testing.T) {
	d := NewScalar(T(42))
	assertDenseDims(t, 1, 1, d)
	assert.Equal(t, []T{42}, d.Data())
}

func TestNewEmptyVecDense(t *testing.T) {
	t.Run("float32", testNewEmptyVecDense[float32])
	t.Run("float64", testNewEmptyVecDense[float64])
}

func testNewEmptyVecDense[T DType](t *testing.T) {
	t.Run("negative size", func(t *testing.T) {
		require.Panics(t, func() {
			NewEmptyVecDense[T](-1)
		})
	})

	for _, size := range []int{0, 1, 2, 10, 100} {
		t.Run(fmt.Sprintf("size %d", size), func(t *testing.T) {
			d := NewEmptyVecDense[T](size)
			assertDenseDims(t, size, 1, d)
			for _, v := range d.Data() {
				require.Equal(t, T(0), v)
			}
		})
	}
}

func TestNewEmptyDense(t *testing.T) {
	t.Run("float32", testNewEmptyDense[float32])
	t.Run("float64", testNewEmptyDense[float64])
}

func testNewEmptyDense[T DType](t *testing.T) {
	t.Run("negative rows", func(t *testing.T) {
		require.Panics(t, func() {
			NewEmptyDense[T](-1, 1)
		})
	})

	t.Run("negative cols", func(t *testing.T) {
		require.Panics(t, func() {
			NewEmptyDense[T](1, -1)
		})
	})

	for _, r := range []int{0, 1, 2, 10, 100} {
		for _, c := range []int{0, 1, 2, 10, 100} {
			t.Run(fmt.Sprintf("%d x %d", r, c), func(t *testing.T) {
				d := NewEmptyDense[T](r, c)
				assertDenseDims(t, r, c, d)
				for _, v := range d.Data() {
					require.Equal(t, T(0), v)
				}
			})
		}
	}
}

func TestNewOneHotVecDense(t *testing.T) {
	t.Run("float32", testNewOneHotVecDense[float32])
	t.Run("float64", testNewOneHotVecDense[float64])
}

func testNewOneHotVecDense[T DType](t *testing.T) {
	t.Run("negative size", func(t *testing.T) {
		require.Panics(t, func() {
			NewOneHotVecDense[T](-1, 0)
		})
	})

	t.Run("zero size", func(t *testing.T) {
		require.Panics(t, func() {
			NewOneHotVecDense[T](0, 0)
		})
	})

	t.Run("oneAt >= size", func(t *testing.T) {
		require.Panics(t, func() {
			NewOneHotVecDense[T](1, 1)
		})
	})

	t.Run("oneAt negative", func(t *testing.T) {
		require.Panics(t, func() {
			NewOneHotVecDense[T](1, -1)
		})
	})

	testCases := []struct {
		s int
		i int
		d []T
	}{
		{1, 0, []T{1}},
		{2, 0, []T{1, 0}},
		{2, 1, []T{0, 1}},
		{3, 0, []T{1, 0, 0}},
		{3, 1, []T{0, 1, 0}},
		{3, 2, []T{0, 0, 1}},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d, %d", tc.s, tc.i), func(t *testing.T) {
			d := NewOneHotVecDense[T](tc.s, tc.i)
			assertDenseDims(t, tc.s, 1, d)
			assert.Equal(t, tc.d, d.Data())
		})
	}
}

func TestNewInitDense(t *testing.T) {
	t.Run("float32", testNewInitDense[float32])
	t.Run("float64", testNewInitDense[float64])
}

func testNewInitDense[T DType](t *testing.T) {
	t.Run("negative rows", func(t *testing.T) {
		require.Panics(t, func() {
			NewInitDense(-1, 1, T(42))
		})
	})

	t.Run("negative cols", func(t *testing.T) {
		require.Panics(t, func() {
			NewInitDense(1, -1, T(42))
		})
	})

	for _, r := range []int{0, 1, 2, 10, 100} {
		for _, c := range []int{0, 1, 2, 10, 100} {
			t.Run(fmt.Sprintf("%d x %d", r, c), func(t *testing.T) {
				d := NewInitDense(r, c, T(42))
				assertDenseDims(t, r, c, d)
				for _, v := range d.Data() {
					require.Equal(t, T(42), v)
				}
			})
		}
	}
}

func TestNewInitFuncDense(t *testing.T) {
	t.Run("float32", testNewInitFuncDense[float32])
	t.Run("float64", testNewInitFuncDense[float64])
}

func testNewInitFuncDense[T DType](t *testing.T) {
	t.Run("negative rows", func(t *testing.T) {
		require.Panics(t, func() {
			NewInitFuncDense[T](-1, 1, func(r int, c int) T {
				t.Fatal("the callback should not be called")
				return 0
			})
		})
	})

	t.Run("negative cols", func(t *testing.T) {
		require.Panics(t, func() {
			NewInitFuncDense[T](1, -1, func(r int, c int) T {
				t.Fatal("the callback should not be called")
				return 0
			})
		})
	})

	testCases := []struct {
		r int
		c int
		d []T
	}{
		// Each value is a 2-digit number having the format "<row><col>"
		{0, 0, []T{}},
		{0, 1, []T{}},
		{1, 0, []T{}},
		{1, 1, []T{11}},
		{2, 1, []T{11, 21}},
		{3, 1, []T{11, 21, 31}},
		{1, 3, []T{11, 12, 13}},
		{2, 2, []T{
			11, 12,
			21, 22,
		}},
		{3, 3, []T{
			11, 12, 13,
			21, 22, 23,
			31, 32, 33,
		}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d", tc.r, tc.c), func(t *testing.T) {
			d := NewInitFuncDense[T](tc.r, tc.c, func(r int, c int) T {
				if len(tc.d) == 0 {
					t.Fatal("the callback should not be called")
				}
				return T(c + 1 + (r+1)*10)
			})
			assertDenseDims(t, tc.r, tc.c, d)
			assert.Equal(t, tc.d, d.Data())
		})
	}
}

func TestNewInitVecDense(t *testing.T) {
	t.Run("float32", testNewInitVecDense[float32])
	t.Run("float64", testNewInitVecDense[float64])
}

func testNewInitVecDense[T DType](t *testing.T) {
	t.Run("negative size", func(t *testing.T) {
		require.Panics(t, func() {
			NewInitVecDense(-1, T(42))
		})
	})

	for _, size := range []int{0, 1, 2, 10, 100} {
		t.Run(fmt.Sprintf("size %d", size), func(t *testing.T) {
			d := NewInitVecDense(size, T(42))
			assertDenseDims(t, size, 1, d)
			for _, v := range d.Data() {
				require.Equal(t, T(42), v)
			}
		})
	}
}

func TestNewIdentityDense(t *testing.T) {
	t.Run("float32", testNewIdentityDense[float32])
	t.Run("float64", testNewIdentityDense[float64])
}

func testNewIdentityDense[T DType](t *testing.T) {
	t.Run("negative size", func(t *testing.T) {
		require.Panics(t, func() {
			NewIdentityDense[T](-1)
		})
	})

	testCases := []struct {
		s int
		d []T
	}{
		{0, []T{}},
		{1, []T{1}},
		{2, []T{
			1, 0,
			0, 1,
		}},
		{3, []T{
			1, 0, 0,
			0, 1, 0,
			0, 0, 1,
		}},
		{4, []T{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		}},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("size %d", tc.s), func(t *testing.T) {
			d := NewIdentityDense[T](tc.s)
			assertDenseDims(t, tc.s, tc.s, d)
			assert.Equal(t, tc.d, d.Data())
		})
	}
}

func TestDense_SetData(t *testing.T) {
	t.Run("float32", testDenseSetData[float32])
	t.Run("float64", testDenseSetData[float64])
}

func testDenseSetData[T DType](t *testing.T) {
	t.Run("incompatible data size", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.SetData([]T{1, 2, 3})
		})
	})

	t.Run("zero size - nil", func(t *testing.T) {
		d := NewEmptyDense[T](0, 0)
		d.SetData(nil)
		assert.Equal(t, []T{}, d.data)
	})

	t.Run("zero size - empty slice", func(t *testing.T) {
		d := NewEmptyDense[T](0, 0)
		d.SetData([]T{})
		assert.Equal(t, []T{}, d.data)
	})

	t.Run("data is set correctly", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		v := []T{1, 2, 3, 7, 8, 9}
		d.SetData(v)
		assert.Equal(t, v, d.data)
	})

	t.Run("data is copied", func(t *testing.T) {
		d := NewEmptyDense[T](1, 1)
		s := []T{1}
		d.SetData(s)
		s[0] = 42 // modifying s must not modify d.data
		assert.Equal(t, T(1), d.data[0])
	})
}

func TestDense_ZerosLike(t *testing.T) {
	t.Run("float32", testDenseZerosLike[float32])
	t.Run("float64", testDenseZerosLike[float64])
}

func testDenseZerosLike[T DType](t *testing.T) {
	for _, r := range []int{0, 1, 2, 10, 100} {
		for _, c := range []int{0, 1, 2, 10, 100} {
			t.Run(fmt.Sprintf("%d x %d", r, c), func(t *testing.T) {
				d1 := NewInitDense(r, c, T(42))
				d2 := d1.ZerosLike()
				assertDenseDims(t, r, c, d2.(*Dense[T]))
				for _, v := range d2.Data() {
					require.Equal(t, T(0), v)
				}
			})
		}
	}
}

func TestDense_OnesLike(t *testing.T) {
	t.Run("float32", testDenseOnesLike[float32])
	t.Run("float64", testDenseOnesLike[float64])
}

func testDenseOnesLike[T DType](t *testing.T) {
	for _, r := range []int{0, 1, 2, 10, 100} {
		for _, c := range []int{0, 1, 2, 10, 100} {
			t.Run(fmt.Sprintf("%d x %d", r, c), func(t *testing.T) {
				d1 := NewInitDense(r, c, T(42))
				d2 := d1.OnesLike()
				assertDenseDims(t, r, c, d2.(*Dense[T]))
				for _, v := range d2.Data() {
					require.Equal(t, T(1), v)
				}
			})
		}
	}
}

func TestDense_Scalar(t *testing.T) {
	t.Run("float32", testDenseScalar[float32])
	t.Run("float64", testDenseScalar[float64])
}

func testDenseScalar[T DType](t *testing.T) {
	t.Run("non-scalar matrix", func(t *testing.T) {
		d := NewEmptyDense[T](1, 2)
		require.Panics(t, func() {
			d.Scalar()
		})
	})

	t.Run("scalar matrix", func(t *testing.T) {
		d := NewScalar(T(42))
		require.Equal(t, T(42), d.Scalar())
	})
}

func TestDense_Zeros(t *testing.T) {
	t.Run("float32", testDenseZeros[float32])
	t.Run("float64", testDenseZeros[float64])
}

func testDenseZeros[T DType](t *testing.T) {
	for _, r := range []int{0, 1, 2, 10, 100} {
		for _, c := range []int{0, 1, 2, 10, 100} {
			t.Run(fmt.Sprintf("%d x %d", r, c), func(t *testing.T) {
				d := NewInitDense(r, c, T(42))
				d.Zeros()
				assertDenseDims(t, r, c, d)
				for _, v := range d.Data() {
					require.Equal(t, T(0), v)
				}
			})
		}
	}
}

func TestDense_Set(t *testing.T) {
	t.Run("float32", testDenseSet[float32])
	t.Run("float64", testDenseSet[float64])
}

func testDenseSet[T DType](t *testing.T) {
	t.Run("negative row", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.Set(-1, 1, 42)
		})
	})

	t.Run("negative col", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.Set(1, -1, 42)
		})
	})

	t.Run("row out of upper bound", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.Set(2, 1, 42)
		})
	})

	t.Run("col out of upper bound", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.Set(1, 3, 42)
		})
	})

	testCases := []struct {
		r    int
		c    int
		setR int
		setC int
		d    []T
	}{
		{1, 1, 0, 0, []T{42}},

		{2, 1, 0, 0, []T{42, 0}},
		{2, 1, 1, 0, []T{0, 42}},

		{1, 2, 0, 0, []T{42, 0}},
		{1, 2, 0, 1, []T{0, 42}},

		{2, 2, 0, 0, []T{
			42, 0,
			0, 0,
		}},
		{2, 2, 0, 1, []T{
			0, 42,
			0, 0,
		}},
		{2, 2, 1, 0, []T{
			0, 0,
			42, 0,
		}},
		{2, 2, 1, 1, []T{
			0, 0,
			0, 42,
		}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d set (%d, %d)", tc.r, tc.c, tc.setR, tc.setC), func(t *testing.T) {
			d := NewEmptyDense[T](tc.r, tc.c)
			d.Set(tc.setR, tc.setC, 42)
			assert.Equal(t, tc.d, d.Data())
		})
	}
}

func TestDense_At(t *testing.T) {
	t.Run("float32", testDenseAt[float32])
	t.Run("float64", testDenseAt[float64])
}

func testDenseAt[T DType](t *testing.T) {
	t.Run("negative row", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.At(-1, 1)
		})
	})

	t.Run("negative col", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.At(1, -1)
		})
	})

	t.Run("row out of upper bound", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.At(2, 1)
		})
	})

	t.Run("col out of upper bound", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.At(1, 3)
		})
	})

	testCases := []struct {
		r   int
		c   int
		atR int
		atC int
		v   T
	}{
		// Each value is a 2-digit number having the format "<row><col>"
		{1, 1, 0, 0, 11},

		{2, 1, 0, 0, 11},
		{2, 1, 1, 0, 21},

		{1, 2, 0, 0, 11},
		{1, 2, 0, 1, 12},

		{2, 2, 0, 0, 11},
		{2, 2, 0, 1, 12},
		{2, 2, 1, 0, 21},
		{2, 2, 1, 1, 22},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d at (%d, %d)", tc.r, tc.c, tc.atR, tc.atC), func(t *testing.T) {
			d := NewInitFuncDense[T](tc.r, tc.c, func(r int, c int) T {
				return T(c + 1 + (r+1)*10)
			})
			v := d.At(tc.atR, tc.atC)
			assert.Equal(t, tc.v, v)
		})
	}
}

func TestDense_SetVec(t *testing.T) {
	t.Run("float32", testDenseSetVec[float32])
	t.Run("float64", testDenseSetVec[float64])
}

func testDenseSetVec[T DType](t *testing.T) {
	t.Run("non-vector matrix", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.SetVec(1, 42)
		})
	})

	t.Run("negative index", func(t *testing.T) {
		d := NewEmptyVecDense[T](2)
		require.Panics(t, func() {
			d.SetVec(-1, 42)
		})
	})

	t.Run("index out of upper bound", func(t *testing.T) {
		d := NewEmptyVecDense[T](2)
		require.Panics(t, func() {
			d.SetVec(2, 42)
		})
	})

	testCases := []struct {
		size int
		i    int
		d    []T
	}{
		{1, 0, []T{42}},
		{2, 0, []T{42, 0}},
		{2, 1, []T{0, 42}},
		{4, 0, []T{42, 0, 0, 0}},
		{4, 1, []T{0, 42, 0, 0}},
		{4, 2, []T{0, 0, 42, 0}},
		{4, 3, []T{0, 0, 0, 42}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("row vector size %d set %d", tc.size, tc.i), func(t *testing.T) {
			d := NewEmptyDense[T](tc.size, 1)
			d.SetVec(tc.i, 42)
			assert.Equal(t, tc.d, d.Data())
		})

		t.Run(fmt.Sprintf("column vector size %d set %d", tc.size, tc.i), func(t *testing.T) {
			d := NewEmptyDense[T](1, tc.size)
			d.SetVec(tc.i, 42)
			assert.Equal(t, tc.d, d.Data())
		})
	}
}

func TestDense_AtVec(t *testing.T) {
	t.Run("float32", testDenseAtVec[float32])
	t.Run("float64", testDenseAtVec[float64])
}

func testDenseAtVec[T DType](t *testing.T) {
	t.Run("non-vector matrix", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.AtVec(1)
		})
	})

	t.Run("negative index", func(t *testing.T) {
		d := NewEmptyVecDense[T](2)
		require.Panics(t, func() {
			d.AtVec(-1)
		})
	})

	t.Run("index out of upper bound", func(t *testing.T) {
		d := NewEmptyVecDense[T](2)
		require.Panics(t, func() {
			d.AtVec(2)
		})
	})

	testCases := []struct {
		size int
		i    int
		v    T
	}{
		{1, 0, 1},
		{2, 0, 1},
		{2, 1, 2},
		{4, 0, 1},
		{4, 1, 2},
		{4, 2, 3},
		{4, 3, 4},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("row vector size %d set %d", tc.size, tc.i), func(t *testing.T) {
			d := NewInitFuncDense[T](tc.size, 1, func(r, _ int) T {
				return T(r + 1)
			})
			v := d.AtVec(tc.i)
			assert.Equal(t, tc.v, v)
		})

		t.Run(fmt.Sprintf("column vector size %d set %d", tc.size, tc.i), func(t *testing.T) {
			d := NewInitFuncDense[T](1, tc.size, func(_, c int) T {
				return T(c + 1)
			})
			v := d.AtVec(tc.i)
			assert.Equal(t, tc.v, v)
		})
	}
}

func TestDense_ExtractRow(t *testing.T) {
	t.Run("float32", testDenseExtractRow[float32])
	t.Run("float64", testDenseExtractRow[float64])
}

func testDenseExtractRow[T DType](t *testing.T) {
	t.Run("negative row", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ExtractRow(-1)
		})
	})

	t.Run("row out of upper bound", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ExtractRow(2)
		})
	})

	testCases := []struct {
		r int
		c int
		i int
		d []T
	}{
		// Each value is a 2-digit number having the format "<row><col>"
		{1, 0, 0, []T{}},
		{1, 1, 0, []T{11}},
		{1, 2, 0, []T{11, 12}},

		{2, 1, 0, []T{11}},
		{2, 1, 1, []T{21}},

		{2, 2, 0, []T{11, 12}},
		{2, 2, 1, []T{21, 22}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d extract %d", tc.r, tc.c, tc.i), func(t *testing.T) {
			d := NewInitFuncDense[T](tc.r, tc.c, func(r int, c int) T {
				return T(c + 1 + (r+1)*10)
			})
			r := d.ExtractRow(tc.i)
			assertDenseDims(t, len(tc.d), 1, r.(*Dense[T]))
			assert.Equal(t, tc.d, r.Data())
		})
	}
}

func TestDense_ExtractColumn(t *testing.T) {
	t.Run("float32", testDenseExtractColumn[float32])
	t.Run("float64", testDenseExtractColumn[float64])
}

func testDenseExtractColumn[T DType](t *testing.T) {
	t.Run("negative col", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ExtractColumn(-1)
		})
	})

	t.Run("col out of upper bound", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ExtractColumn(3)
		})
	})

	testCases := []struct {
		r int
		c int
		i int
		d []T
	}{
		// Each value is a 2-digit number having the format "<row><col>"
		{0, 1, 0, []T{}},
		{1, 1, 0, []T{11}},
		{2, 1, 0, []T{11, 21}},

		{1, 2, 0, []T{11}},
		{1, 2, 1, []T{12}},

		{2, 2, 0, []T{11, 21}},
		{2, 2, 1, []T{12, 22}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d extract %d", tc.r, tc.c, tc.i), func(t *testing.T) {
			d := NewInitFuncDense[T](tc.r, tc.c, func(r int, c int) T {
				return T(c + 1 + (r+1)*10)
			})
			c := d.ExtractColumn(tc.i)
			assertDenseDims(t, len(tc.d), 1, c.(*Dense[T]))
			assert.Equal(t, tc.d, c.Data())
		})
	}
}

func TestDense_View(t *testing.T) {
	t.Run("float32", testDenseView[float32])
	t.Run("float64", testDenseView[float64])
}

func testDenseView[T DType](t *testing.T) {
	t.Run("negative rows", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.View(-1, 6)
		})
	})

	t.Run("negative cols", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.View(6, -1)
		})
	})

	t.Run("incompatible dimensions", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.View(2, 2)
		})
	})

	testCases := []struct {
		r     int
		c     int
		viewR int
		viewC int
	}{
		{0, 0, 0, 0},
		{1, 1, 1, 1},

		{0, 1, 0, 1},
		{0, 1, 1, 0},

		{1, 0, 1, 0},
		{1, 0, 0, 1},

		{1, 2, 1, 2},
		{1, 2, 2, 1},

		{2, 1, 2, 1},
		{2, 1, 1, 2},

		{2, 2, 2, 2},

		// Weird cases, but technically legit
		{2, 2, 1, 4},
		{2, 2, 4, 1},
		{2, 3, 2, 3},
		{2, 3, 3, 2},
		{2, 3, 1, 6},
		{2, 3, 6, 1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d view %d x %d", tc.r, tc.c, tc.viewR, tc.viewC), func(t *testing.T) {
			d := NewEmptyDense[T](tc.r, tc.c)
			v := d.View(tc.viewR, tc.viewC)
			assertDenseDims(t, tc.viewR, tc.viewC, v.(*Dense[T]))
			assert.Equal(t, d.Data(), v.Data())
		})
	}

	t.Run("data is not copied", func(t *testing.T) {
		d := NewEmptyDense[T](1, 1)
		v := d.View(1, 1)
		d.Set(0, 0, 42) // modifying d must modify v too
		assert.Equal(t, T(42), v.At(0, 0))
		v.Set(0, 0, 2) // modifying v must modify d too
		assert.Equal(t, T(2), d.At(0, 0))
	})
}

func TestDense_Reshape(t *testing.T) {
	t.Run("float32", testDenseReshape[float32])
	t.Run("float64", testDenseReshape[float64])
}

func testDenseReshape[T DType](t *testing.T) {
	t.Run("negative rows", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.Reshape(-1, 6)
		})
	})

	t.Run("negative cols", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.Reshape(6, -1)
		})
	})

	t.Run("incompatible dimensions", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.Reshape(2, 2)
		})
	})

	testCases := []struct {
		r     int
		c     int
		reshR int
		reshC int
	}{
		{0, 0, 0, 0},
		{1, 1, 1, 1},

		{0, 1, 0, 1},
		{0, 1, 1, 0},

		{1, 0, 1, 0},
		{1, 0, 0, 1},

		{1, 2, 1, 2},
		{1, 2, 2, 1},

		{2, 1, 2, 1},
		{2, 1, 1, 2},

		{2, 2, 2, 2},

		// Weird cases, but technically legit
		{2, 2, 1, 4},
		{2, 2, 4, 1},
		{2, 3, 2, 3},
		{2, 3, 3, 2},
		{2, 3, 1, 6},
		{2, 3, 6, 1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d reshape %d x %d", tc.r, tc.c, tc.reshR, tc.reshC), func(t *testing.T) {
			d := NewEmptyDense[T](tc.r, tc.c)
			r := d.Reshape(tc.reshR, tc.reshC)
			assertDenseDims(t, tc.reshR, tc.reshC, r.(*Dense[T]))
			assert.Equal(t, d.Data(), r.Data())
		})
	}

	t.Run("data is copied", func(t *testing.T) {
		d := NewEmptyDense[T](1, 1)
		r := d.Reshape(1, 1)
		d.Set(0, 0, 42) // modifying d must not modify r
		assert.Equal(t, T(0), r.At(0, 0))
	})
}

func TestDense_ReshapeInPlace(t *testing.T) {
	t.Run("float32", testDenseReshapeInPlace[float32])
	t.Run("float64", testDenseReshapeInPlace[float64])
}

func testDenseReshapeInPlace[T DType](t *testing.T) {
	t.Run("negative rows", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ReshapeInPlace(-1, 6)
		})
	})

	t.Run("negative cols", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ReshapeInPlace(6, -1)
		})
	})

	t.Run("incompatible dimensions", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ReshapeInPlace(2, 2)
		})
	})

	testCases := []struct {
		r     int
		c     int
		reshR int
		reshC int
	}{
		{0, 0, 0, 0},
		{1, 1, 1, 1},

		{0, 1, 0, 1},
		{0, 1, 1, 0},

		{1, 0, 1, 0},
		{1, 0, 0, 1},

		{1, 2, 1, 2},
		{1, 2, 2, 1},

		{2, 1, 2, 1},
		{2, 1, 1, 2},

		{2, 2, 2, 2},

		// Weird cases, but technically legit
		{2, 2, 1, 4},
		{2, 2, 4, 1},
		{2, 3, 2, 3},
		{2, 3, 3, 2},
		{2, 3, 1, 6},
		{2, 3, 6, 1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d reshape %d x %d", tc.r, tc.c, tc.reshR, tc.reshC), func(t *testing.T) {
			d := NewEmptyDense[T](tc.r, tc.c)
			d2 := d.ReshapeInPlace(tc.reshR, tc.reshC)
			assert.Same(t, d, d2)
			assertDenseDims(t, tc.reshR, tc.reshC, d)
		})
	}
}

func TestDense_ResizeVector(t *testing.T) {
	t.Run("float32", testDenseResizeVector[float32])
	t.Run("float64", testDenseResizeVector[float64])
}

func testDenseResizeVector[T DType](t *testing.T) {
	t.Run("non-vector matrix", func(t *testing.T) {
		d := NewEmptyDense[T](2, 3)
		require.Panics(t, func() {
			d.ResizeVector(2)
		})
	})

	t.Run("negative size", func(t *testing.T) {
		d := NewEmptyVecDense[T](2)
		require.Panics(t, func() {
			d.ResizeVector(-1)
		})
	})

	testCases := []struct {
		size    int
		newSize int
		d       []T
	}{
		{0, 0, []T{}},

		{1, 0, []T{}},
		{1, 1, []T{1}},
		{1, 2, []T{1, 0}},
		{1, 3, []T{1, 0, 0}},

		{2, 0, []T{}},
		{2, 1, []T{1}},
		{2, 2, []T{1, 2}},
		{2, 3, []T{1, 2, 0}},
		{2, 4, []T{1, 2, 0, 0}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("row vector size %d resize %d", tc.size, tc.newSize), func(t *testing.T) {
			d := NewInitFuncDense[T](tc.size, 1, func(r, _ int) T {
				return T(r + 1)
			})
			r := d.ResizeVector(tc.newSize)
			assert.Equal(t, tc.d, r.Data())
		})

		t.Run(fmt.Sprintf("column vector size %d resize %d", tc.size, tc.newSize), func(t *testing.T) {
			d := NewInitFuncDense[T](1, tc.size, func(_, c int) T {
				return T(c + 1)
			})
			r := d.ResizeVector(tc.newSize)
			assert.Equal(t, tc.d, r.Data())
		})
	}

	t.Run("data is copied - smaller size", func(t *testing.T) {
		d := NewEmptyVecDense[T](2)
		r := d.ResizeVector(1)
		d.Set(0, 0, 42) // modifying d must not modify r
		assert.Equal(t, T(0), r.At(0, 0))
	})

	t.Run("data is copied - bigger size", func(t *testing.T) {
		d := NewEmptyVecDense[T](2)
		r := d.ResizeVector(3)
		d.Set(0, 0, 42) // modifying d must not modify r
		assert.Equal(t, T(0), r.At(0, 0))
	})
}

func TestDense_T(t *testing.T) {
	t.Run("float32", testDenseT[float32])
	t.Run("float64", testDenseT[float64])
}

func testDenseT[T DType](t *testing.T) {
	testCases := []struct {
		r int
		c int
		d []T
	}{
		// Each value is a 2-digit number having the format "<row><col>"
		{0, 0, []T{}},
		{0, 1, []T{}},
		{1, 0, []T{}},
		{1, 1, []T{11}},
		{1, 2, []T{11, 12}},
		{2, 1, []T{11, 21}},
		{2, 2, []T{
			11, 21,
			12, 22,
		}},
		{2, 3, []T{
			11, 21,
			12, 22,
			13, 23,
		}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d x %d", tc.r, tc.c), func(t *testing.T) {
			d := NewInitFuncDense[T](tc.r, tc.c, func(r int, c int) T {
				return T(c + 1 + (r+1)*10)
			})
			tr := d.T()
			assertDenseDims(t, tc.c, tc.r, tr.(*Dense[T]))
			assert.Equal(t, tc.d, tr.Data())
		})
	}
}

type prodTestCase[T DType] struct {
	a *Dense[T]
	b *Dense[T]
	y []T
}

func prodTestCases[T DType]() []prodTestCase[T] {
	return []prodTestCase[T]{
		{NewEmptyDense[T](0, 0), NewEmptyDense[T](0, 0), []T{}},
		{NewEmptyDense[T](0, 1), NewEmptyDense[T](0, 1), []T{}},
		{NewEmptyDense[T](1, 0), NewEmptyDense[T](1, 0), []T{}},
		{NewDense[T](1, 1, []T{2}), NewDense[T](1, 1, []T{10}), []T{20}},
		{
			NewDense[T](1, 2, []T{2, 3}),
			NewDense[T](1, 2, []T{10, 20}),
			[]T{20, 60},
		},
		{
			NewDense[T](1, 2, []T{2, 3}),   // row vec
			NewDense[T](2, 1, []T{10, 20}), // col vec
			[]T{20, 60},
		},
		{
			NewDense[T](2, 3, []T{
				2, 3, 4,
				5, 6, 7,
			}),
			NewDense[T](2, 3, []T{
				10, 20, 30,
				40, 50, 60,
			}),
			[]T{
				20, 60, 120,
				200, 300, 420,
			},
		},
	}
}

func TestDense_Prod(t *testing.T) {
	t.Run("float32", testDenseProd[float32])
	t.Run("float64", testDenseProd[float64])
}

func testDenseProd[T DType](t *testing.T) {
	t.Run("incompatible data size", func(t *testing.T) {
		a := NewEmptyDense[T](2, 3)
		b := NewEmptyDense[T](2, 4)
		require.Panics(t, func() {
			a.Prod(b)
		})
	})

	for _, tc := range prodTestCases[T]() {
		t.Run(fmt.Sprintf("%d x %d, %d x %d", tc.a.rows, tc.a.cols, tc.b.rows, tc.b.cols), func(t *testing.T) {
			y := tc.a.Prod(tc.b)
			assertDenseDims(t, tc.a.rows, tc.a.cols, y.(*Dense[T]))
			assert.Equal(t, tc.y, y.Data())
		})
	}
}

func TestDense_ProdInPlace(t *testing.T) {
	t.Run("float32", testDenseProdInPlace[float32])
	t.Run("float64", testDenseProdInPlace[float64])
}

func testDenseProdInPlace[T DType](t *testing.T) {
	t.Run("incompatible data size", func(t *testing.T) {
		a := NewEmptyDense[T](2, 3)
		b := NewEmptyDense[T](2, 4)
		require.Panics(t, func() {
			a.ProdInPlace(b)
		})
	})

	for _, tc := range prodTestCases[T]() {
		t.Run(fmt.Sprintf("%d x %d, %d x %d", tc.a.rows, tc.a.cols, tc.b.rows, tc.b.cols), func(t *testing.T) {
			a2 := tc.a.ProdInPlace(tc.b)
			assert.Same(t, tc.a, a2)
			assert.Equal(t, tc.y, tc.a.Data())
		})
	}
}

func TestDense_AddScalar(t *testing.T) {
	t.Run("float32", testDenseAddScalar[float32])
	t.Run("float64", testDenseAddScalar[float64])
}

func testDenseAddScalar[T DType](t *testing.T) {
	// TODO: this is just a quick test; test corner cases...
	a := NewVecDense([]T{1, 2, 3, 0})
	b := a.AddScalar(10)
	assertSliceEqualApprox(t, []T{11, 12, 13, 10}, b.Data())
}

func assertDenseDims[T DType](t *testing.T, expectedRows, expectedCols int, d *Dense[T]) {
	t.Helper()

	expectedSize := expectedRows * expectedCols
	dimsRows, dimsCols := d.Dims()

	assert.NotNil(t, d)
	assert.Equal(t, expectedRows, d.Rows())
	assert.Equal(t, expectedRows, dimsRows)
	assert.Equal(t, expectedCols, d.Columns())
	assert.Equal(t, expectedCols, dimsCols)
	assert.Equal(t, expectedSize, d.Size())
	assert.Len(t, d.Data(), expectedSize)
}

func assertSliceEqualApprox[T DType](t *testing.T, expected, actual []T) {
	t.Helper()
	assert.InDeltaSlice(t, expected, actual, 1.0e-04)
}
