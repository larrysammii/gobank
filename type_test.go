package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	acc, err := NewAccount("Penis", "Cock", "shittyPassword69")
	assert.Nil(t, err)

	fmt.Printf("%+v\n", acc)
}
