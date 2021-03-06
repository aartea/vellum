//  Copyright (c) 2017 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vellum

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
)

type registry struct {
	table     []*builderState
	tableSize uint
	mruSize   uint
	hasher    hash.Hash64
}

func newRegistry(tableSize, mruSize int) *registry {
	nsize := tableSize * mruSize
	rv := &registry{
		table:     make([]*builderState, nsize),
		tableSize: uint(tableSize),
		mruSize:   uint(mruSize),
		hasher:    fnv.New64a(),
	}
	return rv
}

func (r *registry) entry(node *builderState) *builderState {
	if len(r.table) == 0 {
		return nil
	}
	bucket := r.hash(node)
	start := r.mruSize * uint(bucket)
	end := start + r.mruSize
	rc := registryCache(r.table[start:end])
	return rc.entry(node)
}

func (r *registry) hash(b *builderState) int {
	r.hasher.Reset()
	var final uint64
	if b.final {
		final = 1
	}
	_ = binary.Write(r.hasher, binary.LittleEndian, final)
	_ = binary.Write(r.hasher, binary.LittleEndian, b.finalVal)
	for _, t := range b.transitions {
		_ = binary.Write(r.hasher, binary.LittleEndian, t.key)
		_ = binary.Write(r.hasher, binary.LittleEndian, t.val)
		_ = binary.Write(r.hasher, binary.LittleEndian, t.dest)
	}
	return int(uint(r.hasher.Sum64()) % r.tableSize)
}

type registryCache []*builderState

func (r registryCache) entry(node *builderState) *builderState {
	if len(r) == 1 {
		cell := r[0]
		if cell != nil && cell.equiv(node) {
			return cell
		}
		r[0] = node
		return nil
	}
	for i, ent := range r {
		if ent != nil && ent.equiv(node) {
			r.promote(i)
			return ent
		}
	}
	// no match
	last := len(r) - 1
	r[last] = node // discard LRU
	r.promote(last)
	return nil

}

func (r registryCache) promote(i int) {
	for i > 0 {
		r.swap(i-1, i)
		i--
	}
}

func (r registryCache) swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
