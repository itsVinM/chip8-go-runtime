/*
V. Mocanu
Barcelona
March 2026
*/

package chip8

import (
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	keyCount      uint = 16
	memorySize    uint = 4096
	registerCount uint = 16
	stackLevels   uint = 16
	videoHeight   uint = 32
	videoWidth    uint = 64
)

type Chip8 struct {
	// Public hardware buffers
	Keypad [keyCount]uint8
	Video  [videoHeight * videoWidth]uint64

	// Internal chip state and memory
	memory     [memorySize]uint8
	registers  [registerCount]uint8
	index      uint16
	pc         uint16
	delayTimer uint8
	stack      [stackLevels]uint16
	sp         uint8
	opcode     uint16
	Halt       bool
	LastError  string
}

// Timer - no sound
func (cpu *Chip8) UpdateTimers() {
	if cpu.delayTimer > 0 {
		cpu.delayTimer--
	}
}

func (cpu *Chip8) LoadROM(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return cpu.LoadFromBytes(data)
}

func (cpu *Chip8) LoadFromBytes(data []byte) error {
	// Standard CHIP-8 ROMs start at 0x200
	const offset = 0x200
	if len(data) > (4096 - offset) {
		return fmt.Errorf("ROM too large: %d bytes (max %d)", len(data), 4096-offset)
	}

	cpu.pc = offset
	copy(cpu.memory[offset:], data)
	return nil
}

func New() *Chip8 {
	cpu := &Chip8{
		pc: 0x200, //Start
	}
	// From font.go load the "interpreter area" (0x000-0x050)
	copy(cpu.memory[:], FontSet[:])
	return cpu
}

func (cpu *Chip8) Cycle() {
	// combine 2 bytes from memory into a single 16-bit code
	cpu.opcode = uint16(cpu.memory[cpu.pc])<<8 | uint16(cpu.memory[cpu.pc+1])
	cpu.pc += 2

	// decode variables used by opcode
	x := (cpu.opcode & 0x0F00) >> 8
	y := (cpu.opcode & 0x00F0) >> 4
	n := byte(cpu.opcode & 0x000F)
	kk := byte(cpu.opcode & 0x00FF)
	nnn := cpu.opcode & 0x0FFF

	// Execute
	switch cpu.opcode & 0xF000 {
	case 0x0000:
		switch cpu.opcode & 0x00FF {
		case 0x00E0:
			cpu.op00E0() //CLS
		case 0x00EE:
			cpu.op00EE() // RET
		}

	case 0x1000:
		cpu.pc = nnn // JP address
	case 0x2000:
		cpu.op2nnn(nnn) // CALL address
	case 0x3000:
		cpu.op3xkk(x, kk) // SE Vx, byte
	case 0x4000:
		cpu.op4xkk(x, kk) // SNE Vx, byte
	case 0x5000:
		cpu.op5xy0(x, y) // SE Vx, Vy
	case 0x6000:
		cpu.registers[x] = kk // LD Vx, byte
	case 0x7000:
		cpu.registers[x] += kk // ADD Vx, byte

	case 0x8000:
		switch n {
		case 0x0:
			cpu.registers[x] = cpu.registers[y]
		case 0x1:
			cpu.registers[x] |= cpu.registers[y]
		case 0x2:
			cpu.registers[x] &= cpu.registers[y]
		case 0x3:
			cpu.registers[x] ^= cpu.registers[y]
		case 0x4:
			cpu.op8xy4(x, y) // ADD with carry
		case 0x5:
			cpu.op8xy5(x, y) // SUB
		case 0x6:
			cpu.op8xy6(x) // SHR
		case 0x7:
			cpu.op8xy7(x, y) // SUBN
		case 0xE:
			cpu.op8xyE(x) // SHL
		}
	case 0x9000:
		cpu.op9xy0(x, y) // SNE Vx, Vy
	case 0xA000:
		cpu.index = nnn // LD I, addr
	case 0xB000:
		cpu.pc = nnn + uint16(cpu.registers[0])
	case 0xC000:
		cpu.registers[x] = uint8(rand.Uint32()) & kk
	case 0xD000:
		cpu.opDxyn(x, y, n) // DRW Vx, Vy, height - DRAW
	case 0xE000:
		switch kk {
		case 0x9E:
			cpu.opEx9E(x) // SKP
		case 0xA1:
			cpu.opExA1(x) // SKNP
		}
	case 0xF000:
		cpu.opTableF(x, kk)
	default:
		cpu.Halt = true
		cpu.LastError = fmt.Sprintf("Unknown Opcode: %04X", cpu.opcode)
		fmt.Printf("SYSTEM FAULT: %s at PC: %03X\n", cpu.LastError, cpu.pc)
	}
}

// Instruction implementation
func (cpu *Chip8) op00E0() {
	for i := range cpu.Video {
		cpu.Video[i] = 0
	}
}

func (cpu *Chip8) op00EE() {
	cpu.sp--
	cpu.pc = cpu.stack[cpu.sp]
}

func (cpu *Chip8) op2nnn(nnn uint16) {
	cpu.stack[cpu.sp] = cpu.pc
	cpu.sp++
	cpu.pc = nnn
}

func (cpu *Chip8) op3xkk(x uint16, kk byte) {
	if cpu.registers[x] == kk {
		cpu.pc += 2
	}
}

func (cpu *Chip8) op4xkk(x uint16, kk byte) {
	// SNE Vx, byte
	if cpu.registers[x] != kk {
		cpu.pc += 2
	}
}
func (cpu *Chip8) op5xy0(x, y uint16) {
	// SE Vx, Vy
	if cpu.registers[x] == cpu.registers[y] {
		cpu.pc += 2
	}
}

func (cpu *Chip8) op9xy0(x, y uint16) {
	// SNE Vx, Vy
	if cpu.registers[x] != cpu.registers[y] {
		cpu.pc += 2
	}
}

func (cpu *Chip8) op8xy4(x, y uint16) {
	// ADD with carry
	sum := uint16(cpu.registers[x]) + uint16(cpu.registers[y])
	cpu.registers[0xF] = 0
	if sum > 255 {
		cpu.registers[0xF] = 1
	}
	cpu.registers[x] = uint8(sum & 0xFF)
}

func (cpu *Chip8) op8xy5(x uint16, y uint16) {
	// SUB
	cpu.registers[0xF] = 1
	if cpu.registers[x] < cpu.registers[y] {
		cpu.registers[0xF] = 0
	}
	cpu.registers[x] -= cpu.registers[y]

}

func (cpu *Chip8) op8xy6(x uint16) {
	// SHR - switch to right
	cpu.registers[0xF] = cpu.registers[x] & 0x1
	cpu.registers[x] >>= 1
}

func (cpu *Chip8) op8xy7(x uint16, y uint16) {
	// SUBN
	cpu.registers[0xF] = 1
	if cpu.registers[x] < cpu.registers[y] {
		cpu.registers[0xF] = 0
	}
	cpu.registers[x] = cpu.registers[y] - cpu.registers[x]
}

func (cpu *Chip8) op8xyE(x uint16) {
	// SHL
	cpu.registers[0xF] = (cpu.registers[x] & 0x80) >> 7
	cpu.registers[x] <<= 1
}

func (cpu *Chip8) opEx9E(x uint16) {
	// SKP
	if cpu.Keypad[cpu.registers[x]] != 0 {
		cpu.pc += 2
	}
}
func (cpu *Chip8) opExA1(x uint16) {
	// SKNP
	if cpu.Keypad[cpu.registers[x]] == 0 {
		cpu.pc += 2
	}
}

func (cpu *Chip8) opDxyn(vx, vy uint16, n byte) {
	// calculate target memory range
	if uint32(cpu.index)+uint32(n) > 4096 {
		cpu.Halt = true
		cpu.LastError = "MEM FAULT: Sprite OOB"
		return
	}

	// xPos and yPos are coordinates on the 64x32 grid
	xPos := uint(cpu.registers[vx]) % videoWidth
	yPos := uint(cpu.registers[vy]) % videoHeight
	cpu.registers[0xf] = 0

	for row := uint16(0); row < uint16(n); row++ {
		// Vertical boundary check
		targetY := yPos + uint(row)
		if targetY >= videoHeight {
			break
		}

		// Fetch sprite byte and shift it to the X position
		spriteByte := uint64(cpu.memory[cpu.index+row])

		// Shift by (64 - 8 - xPos) to align the 8-bit sprite into the 64-bit row
		shiftAmount := 64 - 8 - xPos
		spriteRow := spriteByte << shiftAmount

		// Collision Check:
		// If the bits in the current Video row overlap with our new spriteRow bits
		if (cpu.Video[targetY] & spriteRow) != 0 {
			cpu.registers[0xf] = 1
		}

		// XOR the entire row: toggles the bits
		cpu.Video[targetY] ^= spriteRow
	}
}

// mmu
func (cpu *Chip8) opTableF(x uint16, kk byte) {
	switch kk {
	case 0x07:
		cpu.registers[x] = cpu.delayTimer
	case 0x15:
		cpu.delayTimer = cpu.registers[x]
	case 0x1E:
		cpu.index += uint16(cpu.registers[x])
	case 0x29:
		// Fx29 - LD F, Vx: Set I = location of font sprite for character in Vx
		cpu.index = uint16(cpu.registers[x]) * 5
	case 0x33: // BCD - Binary Coded Decimal
		cpu.memory[cpu.index] = cpu.registers[x] / 100
		cpu.memory[cpu.index+1] = (cpu.registers[x] / 10) % 10
		cpu.memory[cpu.index+2] = cpu.registers[x] % 10
	case 0x55:
		for i := uint16(0); i <= x; i++ {
			cpu.memory[cpu.index+i] = cpu.registers[i]
		}
	case 0x65:
		for i := uint16(0); i <= x; i++ {
			cpu.registers[i] = cpu.memory[cpu.index+i]
		}
	case 0x0A:
		// Fx0A - LD Vx, K: Wait for a key press, store the value in Vx.
		paused := true
		for i, key := range cpu.Keypad {
			if key != 0 {
				cpu.registers[x] = uint8(i)
				paused = false
				break
			}
		}

		if paused {
			// Decrement PC by 2 to "repeat" this instruction on the next cycle
			cpu.pc -= 2
		}
	}
}

// ================================================
// DEBUGGING TOOLS
// ================================================

func (debug *Chip8) DrawDebugger(screen *ebiten.Image, x, y int, isPaused bool) {

	// Now you can use isPaused inside the debugger if you want
	statusText := "ACTIVE"
	if isPaused {
		statusText = "HALTED"
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("▼ CPU STATE: %s", statusText), x, y)

	// Direct access to private fields
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PC: 0x%03X", debug.pc), x, y+25)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("I:  0x%03X", debug.index), x, y+40)

	// Draw Registers
	for i, val := range debug.registers {
		row := i / 2
		col := i % 2
		msg := fmt.Sprintf("V%X:%02X", i, val)
		ebitenutil.DebugPrintAt(screen, msg, x+(col*70), y+60+(row*15))
	}

	// Fault Monitoring: Highlight if stack is deep
	if debug.sp > 12 {
		ebitenutil.DebugPrintAt(screen, "!! STACK WARNING !!", x, y+200)
	}
}

// GetCurrentInstruction decodes the current opcode into a string.
func (c *Chip8) GetCurrentInstruction() string {
	if int(c.pc)+1 >= len(c.memory) {
		return "EOF"
	}

	op := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])
	nibble := (op & 0xF000) >> 12
	x := (op & 0x0F00) >> 8
	y := (op & 0x00F0) >> 4
	nnn := op & 0x0FFF
	kk := byte(op & 0x00FF)

	switch nibble {
	case 0x0:
		if op == 0x00E0 {
			return "CLS"
		}
		if op == 0x00EE {
			return "RET"
		}
		return fmt.Sprintf("SYS %03X", nnn)
	case 0x1:
		return fmt.Sprintf("JP  %03X", nnn)
	case 0x2:
		return fmt.Sprintf("CALL %03X", nnn)
	case 0x3:
		return fmt.Sprintf("SE  V%X, %02X", x, kk)
	case 0x4:
		return fmt.Sprintf("SNE V%X, %02X", x, kk)
	case 0x5:
		return fmt.Sprintf("SE  V%X, V%X", x, y)
	case 0x6:
		return fmt.Sprintf("LD  V%X, %02X", x, kk)
	case 0x7:
		return fmt.Sprintf("ADD V%X, %02X", x, kk)
	case 0x8:
		sub := op & 0x000F
		names := map[uint16]string{0: "LD", 1: "OR", 2: "AND", 3: "XOR", 4: "ADD", 5: "SUB", 6: "SHR", 7: "SUBN", 0xE: "SHL"}
		return fmt.Sprintf("%s V%X, V%X", names[sub], x, y)
	case 0xA:
		return fmt.Sprintf("LD  I, %03X", nnn)
	case 0xC:
		return fmt.Sprintf("RND V%X, %02X", x, kk)
	case 0xD:
		return fmt.Sprintf("DRW V%X, V%X, %X", x, y, op&0x000F)
	case 0xE:
		if kk == 0x9E {
			return fmt.Sprintf("SKP V%X", x)
		}
		if kk == 0xA1 {
			return fmt.Sprintf("SKNP V%X", x)
		}
	case 0xF:
		// Map common F-codes
		return fmt.Sprintf("F-OP V%X (%02X)", x, kk)
	}
	return fmt.Sprintf("HEX %04X", op)
}
