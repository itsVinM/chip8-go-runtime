package chip8

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const scale = 15

type PPU struct {
	chip8 *Chip8
}

func NewPPU(chip8 *Chip8) *PPU { return &PPU{chip8: chip8} }

// Debugger
func (p *PPU) drawTileData(screen *ebiten.Image) {
	offsetX, offsetY := 450, 50
	ebitenutil.DebugPrintAt(screen, "▼ TileData", offsetX, offsetY-15)

	// Draw first 128 bytes of memory (Fonts + start of ROM)
	for i := 0; i < 128; i++ {
		b := p.chip8.memory[i]
		gridX := (i % 8) * 10 // 8 sprites per row
		gridY := (i / 8) * 10

		for bit := 0; bit < 8; bit++ {
			if (b >> (7 - bit) & 1) == 1 {
				p.drawPoint(screen, offsetX+gridX+bit, offsetY+gridY, color.White)
			}
		}
	}
}

func (p *PPU) drawPoint(screen *ebiten.Image, x, y int, c color.Color) {
	screen.Set(x, y, c)
}

func (game *PPU) drawGame(screen *ebiten.Image) {
	white := color.RGBA{255, 255, 255, 255}
	screen.Fill(color.Black)

	for y, row := range game.chip8.Video {
		for x := 0; x < 64; x++ {
			// Extract bit from the uint64 (63 down to 0)
			if (row>>(63-x))&1 == 1 {
				// Drawing a rectangle manually
				for sy := 0; sy < scale; sy++ {
					for sx := 0; sx < scale; sx++ {
						screen.Set(x*scale+sx, y*scale+sy, white)
					}
				}
			}
		}
	}
}

func (p *PPU) drawDisassembler(screen *ebiten.Image) {
	offsetX, offsetY := 450, 250 // Position it below the TileData
	ebitenutil.DebugPrintAt(screen, "▼ Disassembler", offsetX, offsetY-15)

	pc := int(p.chip8.pc)

	// We look at a window of 8 instructions (16 bytes)
	for i := -2; i < 6; i++ {
		addr := pc + (i * 2)

		// Stay within memory bounds
		if addr < 0 || addr >= 4094 {
			continue
		}

		// Fetch the 16-bit opcode
		opcode := uint16(p.chip8.memory[addr])<<8 | uint16(p.chip8.memory[addr+1])

		// Decode the opcode into a string
		decoded := p.disassemble(opcode)

		line := fmt.Sprintf("0x%03X: %04X %s", addr, opcode, decoded)

		yPos := offsetY + (i * 15)
		if i == 0 {
			// Highlight the current instruction (the one about to execute)
			ebitenutil.DebugPrintAt(screen, "> "+line, offsetX, yPos)
		} else {
			// Draw surrounding instructions in a dimmer or standard format
			ebitenutil.DebugPrintAt(screen, "  "+line, offsetX, yPos)
		}
	}
}

// disassemble translates raw hex into human-readable Mnemonics
func (p *PPU) disassemble(op uint16) string {
	nnn := op & 0x0FFF
	n := op & 0x000F
	x := (op & 0x0F00) >> 8
	y := (op & 0x00F0) >> 4
	kk := uint8(op & 0x00FF)

	switch op & 0xF000 {
	case 0x0000:
		if op == 0x00E0 {
			return "CLS"
		}
		if op == 0x00EE {
			return "RET"
		}
		return "SYS"
	case 0x1000:
		return fmt.Sprintf("JP   0x%03X", nnn)
	case 0x2000:
		return fmt.Sprintf("CALL 0x%03X", nnn)
	case 0x3000:
		return fmt.Sprintf("SE   V%X, %02X", x, kk)
	case 0x4000:
		return fmt.Sprintf("SNE  V%X, %02X", x, kk)
	case 0x6000:
		return fmt.Sprintf("LD   V%X, %02X", x, kk)
	case 0x7000:
		return fmt.Sprintf("ADD  V%X, %02X", x, kk)
	case 0xA000:
		return fmt.Sprintf("LD   I, 0x%03X", nnn)
	case 0xD000:
		return fmt.Sprintf("DRW  V%X, V%X, %d", x, y, n)
	case 0xE000:
		if kk == 0x9E {
			return fmt.Sprintf("SKP  V%X", x)
		}
		if kk == 0xA1 {
			return fmt.Sprintf("SKNP V%X", x)
		}
	case 0xF000:
		switch kk {
		case 0x07:
			return fmt.Sprintf("LD   V%X, DT", x)
		case 0x15:
			return fmt.Sprintf("LD   DT, V%X", x)
		case 0x18:
			return fmt.Sprintf("LD   ST, V%X", x)
		case 0x29:
			return fmt.Sprintf("LD   F, V%X", x)
		case 0x33:
			return fmt.Sprintf("LD   B, V%X", x)
		case 0x65:
			return fmt.Sprintf("LD   V%X, [I]", x)
		}
	}
	return "???"
}

func (p *PPU) Draw(screen *ebiten.Image) {
	p.drawGame(screen)
	p.drawTileData(screen)
	p.drawDisassembler(screen)
}
