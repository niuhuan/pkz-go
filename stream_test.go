package pkz

import (
	"encoding/json"
	"testing"
)

func TestByteOX(t *testing.T) {
	b := byte(3)
	println("%v", b)
	b ^= key
	println("%v", b)
	b ^= key
	println("%v", b)
}

func TestByteOXArray(t *testing.T) {
	p := []byte{
		1, 2, 3,
	}
	buff, _ := json.Marshal(p)
	println(string(buff))
	for i := 0; i < len(p); i++ {
		p[i] ^= key
	}
	buff, _ = json.Marshal(p)
	println(string(buff))
	for i := 0; i < len(p); i++ {
		p[i] ^= key
	}
	buff, _ = json.Marshal(p)
	println(string(buff))
}
