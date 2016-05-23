package csvutil

import (
	"testing"
)

type Person struct {
	Name    string `csv:"Name"`
	Age     int    `csv:"Age"`
	Address string `csv:"Address"`
	Weight  uint64 `csv:"Weight"`
}

func TestReadFile(t *testing.T) {
	var data []Person

	ReadFile("Person.csv", &data)

	t.Log(data)
	if len(data) == 0 {
		t.Error("read failed.")
	}

}
