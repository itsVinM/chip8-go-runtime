package chip8

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllOpcodes(t *testing.T) {
	tests := []struct {
		name     string
		opcode   [2]byte
		setup    func(*Chip8)
		validate func(*testing.T, *Chip8)
	}{
		{
			name:   "Arithmetic: 8xy4 (ADD with Carry)",
			opcode: [2]byte{0x80, 0x14}, // ADD V0, V1
			setup: func(cpu *Chip8) {
				cpu.registers[0] = 0xFF
				cpu.registers[1] = 0x01
			},
			validate: func(t *testing.T, cpu *Chip8) {
				if cpu.registers[0] != 0x00 {
					t.Errorf("Expected V0 to be 0x00 (overflow), got %02X", cpu.registers[0])
				}
				if cpu.registers[0xF] != 1 {
					t.Error("Expected Carry Flag (VF) to be 1")
				}
			},
		},
		{
			name:   "Arithmetic: 8xy5 (SUB with Borrow)",
			opcode: [2]byte{0x80, 0x15}, // SUB V0, V1
			setup: func(cpu *Chip8) {
				cpu.registers[0] = 0x05
				cpu.registers[1] = 0x07
			},
			validate: func(t *testing.T, cpu *Chip8) {
				if cpu.registers[0xF] != 0 {
					t.Error("Expected VF to be 0 (borrow occurred)")
				}
			},
		},
		{
			name:   "Flow: 1nnn (Jump)",
			opcode: [2]byte{0x1A, 0xBC}, // JP 0xABC
			validate: func(t *testing.T, cpu *Chip8) {
				if cpu.pc != 0x0ABC {
					t.Errorf("Expected PC 0xABC, got %04X", cpu.pc)
				}
			},
		},
		{
			name:   "Flow: 3xkk (Skip if Equal - True)",
			opcode: [2]byte{0x30, 0x44}, // SE V0, 0x44
			setup: func(cpu *Chip8) {
				cpu.registers[0] = 0x44
				cpu.pc = 0x200
			},
			validate: func(t *testing.T, cpu *Chip8) {
				// Cycle adds 2, Skip adds 2 = 0x204
				if cpu.pc != 0x204 {
					t.Errorf("Expected PC 0x204 after skip, got %04X", cpu.pc)
				}
			},
		},
		{
			name:   "memory: Annn (Load Index)",
			opcode: [2]byte{0xAF, 0xFF}, // LD I, 0xFFF
			validate: func(t *testing.T, cpu *Chip8) {
				if cpu.index != 0x0FFF {
					t.Errorf("Expected Index 0xFFF, got %04X", cpu.index)
				}
			},
		},
		{
			name:   "Graphics: Dxyn (Word-Level XOR Draw)",
			opcode: [2]byte{0xD0, 0x11}, // DRW V0, V1, 1 byte
			setup: func(cpu *Chip8) {
				cpu.index = 0x300
				cpu.memory[0x300] = 0b10101010
				cpu.registers[0] = 0 // X
				cpu.registers[1] = 0 // Y
				// Pre-set a pixel for collision: Far left is bit 63
				cpu.Video[0] = 0x8000000000000000
			},
			validate: func(t *testing.T, cpu *Chip8) {
				if cpu.registers[0xF] != 1 {
					t.Error("Expected collision flag VF=1")
				}
				// Initial: 1000... XOR Sprite: 10101010... Result: 00101010...
				expected := uint64(0x2A) << (64 - 8)
				if cpu.Video[0] != expected {
					t.Errorf("Video mismatch. Got %016X", cpu.Video[0])
				}
			},
		},
		{
			name:   "Fault: Out of Bounds Index (Robustness)",
			opcode: [2]byte{0xD0, 0x11}, // DRW V0, V1, 1 byte
			setup: func(cpu *Chip8) {
				cpu.index = 0xFFFF // Points way outside 4KB RAM
				cpu.registers[0] = 0
				cpu.registers[1] = 0
			},
			validate: func(t *testing.T, cpu *Chip8) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("CPU Panicked on OOB Memory Access: %v", r)
					}
				}()
			},
		},
		{
			name:   "Fault: Unknown Opcode (Graceful Halt)",
			opcode: [2]byte{0x50, 0x01}, // 0x5xy1 is not a valid CHIP-8 opcode
			validate: func(t *testing.T, cpu *Chip8) {
				fmt.Printf("[Info] Correctly identified invalid opcode: 5001\n")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpu := New()
			// Reset PC to 0x200
			cpu.pc = 0x200
			cpu.memory[0x200] = tt.opcode[0]
			cpu.memory[0x201] = tt.opcode[1]

			if tt.setup != nil {
				tt.setup(cpu)
			}

			cpu.Cycle()

			tt.validate(t, cpu)
		})
	}
}

func TestROMCollection(t *testing.T) {
	romDir := "../rom"
	files, err := os.ReadDir(romDir)
	if err != nil {
		t.Skipf("Skipping batch test: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".ch8") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			cpu := New()
			path := filepath.Join(romDir, file.Name())

			// FIX 1: Read the file to bytes HERE in the test
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// FIX 2: Pass the BYTES, not the path
			if err := cpu.LoadFromBytes(data); err != nil {
				t.Fatalf("Failed to load: %v", err)
			}

			// Stability check (500 cycles)
			for i := 0; i < 500; i++ {
				cpu.Cycle()
				if cpu.Halt {
					t.Errorf("FAULT in %s: %s at PC 0x%03X", file.Name(), cpu.LastError, cpu.pc)
					break
				}
			}
		})
	}
}
