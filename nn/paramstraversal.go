// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nn

import (
	"fmt"
	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/utils"
	"reflect"
	"strings"
	"sync"
)

// paramsTraversal allows the traversal of Model parameters.
// The given callback is invoked for each parameter of the Model.
// If exploreSubModels is true, every nested Model and its parameters are
// also visited.
type paramsTraversal[T mat.DType] struct {
	callback         func(param Param[T])
	exploreSubModels bool
}

// newParamsTraversal returns a new paramsTraversal.
func newParamsTraversal[T mat.DType](callback func(param Param[T]), exploreSubModels bool) paramsTraversal[T] {
	return paramsTraversal[T]{
		callback:         callback,
		exploreSubModels: exploreSubModels,
	}
}

// walk iterates through all the parameters of m.
// TODO: don't loop the field every time, use a lazy initialized "params list" instead
func (pt paramsTraversal[_]) walk(m interface{}) {
	utils.ForEachField(m, func(field interface{}, name string, rTag reflect.StructTag) {
		tag, err := parseModuleFieldTag(rTag.Get("spago"))
		if err != nil {
			panic(err)
		}
		v := reflect.ValueOf(field)
		switch v.Kind() {
		case reflect.Struct, reflect.Ptr, reflect.Interface:
			pt.walkStructOrPtr(field, name, tag)
		case reflect.Slice:
			pt.walkSlice(v, name, tag)
		case reflect.Map:
			pt.walkMap(v, name, tag)
		}
	})
}

func (pt paramsTraversal[T]) walkStructOrPtr(item interface{}, name string, tag moduleFieldTag) {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() != reflect.Struct {
		return
	}
	switch itemT := item.(type) {
	case *BaseParam[T]:
		pt.walkParam(itemT, name, tag)
	case Model[T]:
		if pt.exploreSubModels {
			pt.walk(item)
		}
	case ag.Differentiable[T]:
		_, isModel := itemT.(Model[T])
		if !isModel {
			pt.walk(item)
		}
	case *sync.Map:
		pt.walkSyncMap(itemT, name, tag)
	default:
		return
	}
}

func (pt paramsTraversal[_]) walkSyncMap(i *sync.Map, name string, tag moduleFieldTag) {
	i.Range(func(key, value interface{}) bool {
		switch k := key.(type) {
		case string:
			key = k
		case int:
			key = fmt.Sprintf("%d", k)
		default:
			return false // skip map if the key is not a string or an int
		}

		name := strings.ToLower(fmt.Sprintf("%s.%s", name, key))
		switch reflect.ValueOf(value).Kind() {
		case reflect.Struct, reflect.Ptr, reflect.Interface:
			pt.walkStructOrPtr(value, name, tag)
		default:
			return false // skip
		}
		return true
	})
}

func (pt paramsTraversal[_]) walkSlice(v reflect.Value, name string, tag moduleFieldTag) {
	length := v.Len()
	for i := 0; i < length; i++ {
		p := v.Index(i)
		switch p.Kind() {
		case reflect.Struct, reflect.Ptr, reflect.Interface:
			pt.walkStructOrPtr(p.Interface(), name, tag)
		default:
			return // skip
		}
	}
}

func (pt paramsTraversal[_]) walkMap(v reflect.Value, name string, tag moduleFieldTag) {
	mapRange := v.MapRange()
	for mapRange.Next() {
		key := ""
		switch k := mapRange.Key().Interface().(type) {
		case string:
			key = k
		case int:
			key = fmt.Sprintf("%d", k)
		default:
			return // skip map if the key is not a string or an int
		}

		name := strings.ToLower(fmt.Sprintf("%s.%s", name, key))
		switch mapRange.Value().Kind() {
		case reflect.Struct, reflect.Ptr, reflect.Interface:
			pt.walkStructOrPtr(mapRange.Value().Interface(), name, tag)
		default:
			return // skip
		}
	}
}

func (pt paramsTraversal[T]) walkParam(item *BaseParam[T], name string, tag moduleFieldTag) {
	if item.Name() == "" {
		item.SetName(strings.ToLower(name))
	}
	item.SetType(tag.paramType())
	pt.callback(item)
}
