package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

func TestOperations(t *testing.T) {
	s := NewStore(nil)

	s.Put("foo", "bar", 1)

	v, err := s.Get("foo")

	if err != nil {
		t.Error(err)
	}

	if v == nil || *v != "bar" {
		t.Error("")
	}

	s.Forget("foo")

	v, err = s.Get("foo")

	if err == nil || v != nil {
		t.Error("There should be an error")
	}
}

func TestTtlExpiration(t *testing.T) {
	s := NewStore(nil)

	s.Put("foo", "bar", 1)
	s.Put("bar", "baz", 2)

	<-time.Tick(time.Second)

	v, err := s.Get("foo")

	if err == nil || v != nil {
		t.Error("foo should be removed by ttl")
	}

	v, err = s.Get("bar")

	if err != nil || v == nil {
		t.Error("bar should not be removed by ttl")
	}
}

func TestParallelOperations(t *testing.T) {
	s := NewStore(nil)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		s.Put("foo", "bar", 60)
		wg.Done()
	}()
	go func() {
		s.Put("bar", "baz", 60)
		wg.Done()
	}()

	wg.Wait()

	wg.Add(2)

	go func() {
		v, _ := s.Get("foo")

		if v == nil || *v != "bar" {
			t.Fatal("foo != bar")
		}

		wg.Done()
	}()

	go func() {
		v, _ := s.Get("bar")

		if v == nil || *v != "baz" {
			t.Fatal("bar != baz")
		}

		wg.Done()
	}()

	wg.Wait()
}

func TestPersistence(t *testing.T) {
	p := "./test.json"
	s := NewStore(&p)

	defer os.Remove(p)

	s.Put("foo", "bar", 10)
	s.Put("bar", "baz", 15)

	s.Close()

	if _, err := os.Stat(p); err != nil {
		t.Error("file not exist", err)
	}

	data, _ := ioutil.ReadFile(p)
	m := make(map[string]*item)

	json.Unmarshal(data, &m)

	if len(m) != 2 {
		t.Error("there should be two items")
	}

	// TODO check values
}
