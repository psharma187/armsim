// Filename: computer.go
// Contents: The Computer struct and related methods.

package armsim

import (
	"fmt"
	"log"
	"os"
)

// A Computer holds the RAM, registers, and CPU of the simulated ARM
// architecture. It has methods to allow the loading and execution of a program
// on the simulator.
type Computer struct {
	// A reference to RAM for the simulator
	ram *Memory

	// A reference to the bank of CPU registers,
	// implemented using a standard memory container
	registers *Memory

	// A reference to the CPU for the simulator
	cpu *CPU

	// A simple counter to track number of execution cycles
	step_counter uint64

	// Logger class
	log *log.Logger
}

// Initializes a Computer
//
// Parameters:
//  memSize - a unsigned 32-bit integer specifying the size of the RAM in the
//  computer
//
// Returns:
//  a pointer to the newly created Computer
func NewComputer(memSize uint32) (c *Computer) {
	c = new(Computer)

	// Setup logging
	c.log  = log.New(os.Stdout, "Computer: ", 0)

	// Initialize RAM of memSize
	c.ram = NewMemory(memSize)

	// Initialize a register bank to contain all 16 registers + CPSR
	c.registers = NewMemory(CPSR + 4)

	// Initialize CPU with RAM and registers
	c.cpu = NewCPU(c.ram, c.registers)

	return
}

// Simulates the running of the a computer. It executes the fetch, execute,
// decode cycle until fetch returns false (signifying an instruction of 0x0).
func (c *Computer) Run() {
	for {
		if !c.Step() {
			break
		}
	}
}

// Performs a single execution cycle. Take no parameters and returns a boolean
// signifying if the cycle was completed (a cycle will not complete if the
// instrution fetched is 0x0).
func (c *Computer) Step() bool {
	instruction := c.cpu.Fetch()

	// Don't continue if the instruction is useless
	if instruction == 0x0 {
		return false
	}

	// Not easily testable at the moment
	c.cpu.Decode()
	c.cpu.Execute()

	// Increment step counter
	c.step_counter++

	return true
}

// Builds a three-line status output to debug simulator.
//
// Returns:
//  string contining trace output
func (c *Computer) Trace() (output string) {
	program_counter, _ := c.registers.ReadWord(PC)

	// Build Flags int
	cpsr, _ := c.registers.ReadWord(CPSR)
	flags := ExtractBits(cpsr, N, F) >> F

	output = fmt.Sprintf("%06d %08x %08x %04d\t", c.step_counter, program_counter,
		c.ram.Checksum(), flags)
	for i := 0; i < 16; i++ {
		reg, _ := c.registers.ReadWord(uint32(i))
		output += fmt.Sprintf("%2d=%08x", i, reg)
		if i == 3 || i == 9 {
			output += "\n\t"
		} else if i != 15 {
			output += "\t"
		}
	}
	c.log.Println("*****TRACE*****")
	c.log.Print(output)

	return
}
