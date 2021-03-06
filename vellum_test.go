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
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestRoundTripSimple(t *testing.T) {
	f, err := ioutil.TempFile("", "vellum")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	defer func() {
		err = os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

	b, err := New(f, nil)
	if err != nil {
		t.Fatalf("error creating builder: %v", err)
	}

	err = insertStringMap(b, smallSample)
	if err != nil {
		t.Fatalf("error building: %v", err)
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("err closing: %v", err)
	}

	fst, err := Open(f.Name())
	if err != nil {
		t.Fatalf("error loading set: %v", err)
	}
	defer func() {
		err = fst.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// first check all the expected values
	got := map[string]uint64{}
	itr, err := fst.Iterator(nil, nil)
	for err == nil {
		key, val := itr.Current()
		got[string(key)] = val
		err = itr.Next()
	}
	if err != ErrIteratorDone {
		t.Errorf("iterator error: %v", err)
	}
	if !reflect.DeepEqual(smallSample, got) {
		t.Errorf("expected %v, got: %v", smallSample, got)
	}

	// some additional tests for items that should not exist
	if ok, _ := fst.Contains([]byte("mo")); ok {
		t.Errorf("expected to not contain mo, but did")
	}

	if ok, _ := fst.Contains([]byte("monr")); ok {
		t.Errorf("expected to not contain monr, but did")
	}

	if ok, _ := fst.Contains([]byte("thur")); ok {
		t.Errorf("expected to not contain thur, but did")
	}

	if ok, _ := fst.Contains([]byte("thurp")); ok {
		t.Errorf("expected to not contain thurp, but did")
	}

	if ok, _ := fst.Contains([]byte("tue")); ok {
		t.Errorf("expected to not contain tue, but did")
	}

	if ok, _ := fst.Contains([]byte("tuesd")); ok {
		t.Errorf("expected to not contain tuesd, but did")
	}

	// a few more misc non-existent values to increase coverage
	if ok, _ := fst.Contains([]byte("x")); ok {
		t.Errorf("expected to not contain x, but did")
	}
}

func TestRoundTripThousand(t *testing.T) {
	dataset := thousandTestWords
	randomThousandVals := randomValues(dataset)

	f, err := ioutil.TempFile("", "vellum")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	defer func() {
		err = os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

	b, err := New(f, nil)
	if err != nil {
		t.Fatalf("error creating builder: %v", err)
	}

	err = insertStrings(b, dataset, randomThousandVals)
	if err != nil {
		t.Fatalf("error inserting thousand words: %v", err)
	}
	err = b.Close()
	if err != nil {
		t.Fatalf("error closing builder: %v", err)
	}

	fst, err := Open(f.Name())
	if err != nil {
		t.Fatalf("error loading set: %v", err)
	}
	defer func() {
		err = fst.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// first check all the expected values
	got := map[string]uint64{}
	itr, err := fst.Iterator(nil, nil)
	for err == nil {
		key, val := itr.Current()
		got[string(key)] = val
		err = itr.Next()
	}
	if err != ErrIteratorDone {
		t.Errorf("iterator error: %v", err)
	}

	for i := 0; i < len(dataset); i++ {
		foundVal, ok := got[dataset[i]]
		if !ok {
			t.Fatalf("expected to find key, but didnt: %s", dataset[i])
		}

		if foundVal != randomThousandVals[i] {
			t.Fatalf("expected value %d for key %s, but got %d", randomThousandVals[i], dataset[i], foundVal)
		}

		// now remove it
		delete(got, dataset[i])
	}

	if len(got) != 0 {
		t.Fatalf("expected got map to be empty after checking, still has %v", got)
	}
}

func TestRoundTripEmpty(t *testing.T) {
	f, err := ioutil.TempFile("", "vellum")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	defer func() {
		err = os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

	b, err := New(f, nil)
	if err != nil {
		t.Fatalf("error creating builder: %v", err)
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("error closing: %v", err)
	}

	fst, err := Open(f.Name())
	if err != nil {
		t.Fatalf("error loading set: %v", err)
	}
	defer func() {
		err = fst.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if fst.Len() != 0 {
		t.Fatalf("expected length 0, got %d", fst.Len())
	}

	// first check all the expected values
	got := map[string]uint64{}
	itr, err := fst.Iterator(nil, nil)
	for err == nil {
		key, val := itr.Current()
		got[string(key)] = val
		err = itr.Next()
	}
	if err != ErrIteratorDone {
		t.Errorf("iterator error: %v", err)
	}
	if len(got) > 0 {
		t.Errorf("expected not to see anything, got %v", got)
	}
}

func TestRoundTripEmptyString(t *testing.T) {
	f, err := ioutil.TempFile("", "vellum")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	defer func() {
		err = os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

	b, err := New(f, nil)
	if err != nil {
		t.Fatalf("error creating builder: %v", err)
	}

	err = b.Insert([]byte(""), 0)
	if err != nil {
		t.Fatalf("error inserting empty string")
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("error closing: %v", err)
	}

	fst, err := Open(f.Name())
	if err != nil {
		t.Fatalf("error loading set: %v", err)
	}
	defer func() {
		err = fst.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if fst.Len() != 1 {
		t.Fatalf("expected length 1, got %d", fst.Len())
	}

	// first check all the expected values
	want := map[string]uint64{
		"": 0,
	}
	got := map[string]uint64{}
	itr, err := fst.Iterator(nil, nil)
	for err == nil {
		key, val := itr.Current()
		got[string(key)] = val
		err = itr.Next()
	}
	if err != ErrIteratorDone {
		t.Errorf("iterator error: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expected %v, got: %v", want, got)
	}
}

func TestRoundTripEmptyStringAndOthers(t *testing.T) {
	f, err := ioutil.TempFile("", "vellum")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	defer func() {
		err = os.Remove(f.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

	b, err := New(f, nil)
	if err != nil {
		t.Fatalf("error creating builder: %v", err)
	}

	err = b.Insert([]byte(""), 0)
	if err != nil {
		t.Fatalf("error inserting empty string")
	}
	err = b.Insert([]byte("a"), 0)
	if err != nil {
		t.Fatalf("error inserting empty string")
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("error closing: %v", err)
	}

	fst, err := Open(f.Name())
	if err != nil {
		t.Fatalf("error loading set: %v", err)
	}
	defer func() {
		err = fst.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if fst.Len() != 2 {
		t.Fatalf("expected length 2, got %d", fst.Len())
	}

	// first check all the expected values
	want := map[string]uint64{
		"":  0,
		"a": 0,
	}
	got := map[string]uint64{}
	itr, err := fst.Iterator(nil, nil)
	for err == nil {
		key, val := itr.Current()
		got[string(key)] = val
		err = itr.Next()
	}
	if err != ErrIteratorDone {
		t.Errorf("iterator error: %v", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expected %v, got: %v", want, got)
	}
}
