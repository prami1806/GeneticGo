package main

import (
	"fmt"
	. "github.com/handcraftsman/Interpreter"
)

type call struct {
	blockName string
}

func (c *call) GetType() InstructionType {
	return Call
}

func NewCall(l *field, m *mower, blockName string) *call {
	instr := call{blockName: blockName}
	return &instr
}

func (c *call) GetBlockName() string {
	return c.blockName
}

func (c *call) String() string {
	return fmt.Sprint(c.blockName)
}
