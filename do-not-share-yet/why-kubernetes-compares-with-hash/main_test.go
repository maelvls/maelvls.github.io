package main

import (
	"hash"
	"hash/fnv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/hashstructure"
)

type ComplexStruct struct {
	Name     string
	Age      uint
	Metadata map[string]interface{}
}

var v = ComplexStruct{
	Name: "mitchellh",
	Age:  64,
	Metadata: map[string]interface{}{
		"car":      true,
		"location": "California",
		"siblings": []string{"Bob", "John"},
	},
}

func BenchmarkMitchellhHashstructure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = hashstructure.Hash(v, nil)
	}
}

func BenchmarkKubernetesComputeHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hasher := fnv.New32a()
		DeepHashObject(hasher, v)
		_ = hasher.Sum32()
	}
}

// From https://github.com/kubernetes/kubernetes/blob/7e75a5ef/pkg/util/hash/hash.go#L25-L37
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}
