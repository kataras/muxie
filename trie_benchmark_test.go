package muxie

import (
	"testing"
)

func initTree(tree *Trie) {
	for _, tt := range tests {
		tree.InsertRoute(tt.key, tt.routeName, nil)
	}
}

// go test -run=XXX -v -bench=BenchmarkTrieInsert -count=3
func BenchmarkTrieInsert(b *testing.B) {
	tree := NewTrie()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		initTree(tree)
	}
}

// go test -run=XXX -v -bench=BenchmarkTrieSearch -count=3
func BenchmarkTrieSearch(b *testing.B) {
	tree := NewTrie()
	initTree(tree)
	params := new(paramsWriter)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := range tests {
			for _, req := range tests[i].requests {
				n := tree.Search(req.path, params)
				if n == nil {
					b.Fatalf("%s: node not found\n", req.path)
				}
				params.reset(nil)
			}
		}
	}
}
