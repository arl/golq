package lq

import (
	"fmt"
	"log"
	"testing"
)

func TestDatabase(t *testing.T) {
	fmt.Println("Hello world!")
	log.Println("Hello world!")

	db := CreateDatabase(0, 0, 10, 10, 5, 5)
	db.RemoveAllObjects()

	f := 3.14
	bin := NewClientProxy(f)
	fmt.Println(bin)
}
