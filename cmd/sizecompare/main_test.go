package main

import "testing"

func BenchmarkSerializeToJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		serializeToJSON(metadata)
	}
}

func BenchmarkSerializeToXML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		serializeToXML(metadata)
	}
}

func BenchmarkSerializeToProtoc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		serializeToProtoc(genMetadata)
	}
}
