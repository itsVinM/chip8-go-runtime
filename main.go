package main

import (
	chip8 "chip8/lib"
	"embed" // Import embed
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// 1. Embed the roms folder into the binary
// Make sure your folder is named 'rom' and is in the same directory as main.go
//
//go:embed rom/*.ch8
var romFiles embed.FS

// List of available ROMs for the browser to cycle through
var availableROMs = []string{
	"rom/chip8_logo.ch8",
	"rom/airplane.ch8",
	"rom/tank.ch8",
	"rom/brix.ch8",
	"rom/invaders.ch8",
	"rom/horseyjump.ch8",
	"rom/pong.ch8",
	"rom/ufo.ch8",
}

const (
	scale      = 12
	gameWidth  = 64 * scale
	gameHeight = 32 * scale
	debugWidth = 300
)

type Game struct {
	chip8  *chip8.Chip8
	paused bool // Changed from 'mode' to 'paused' to match your Update logic
}

var keyMap = map[ebiten.Key]uint8{

	ebiten.Key1: 0x1, ebiten.Key2: 0x2, ebiten.Key3: 0x3, ebiten.Key4: 0xC,
	ebiten.KeyQ: 0x4, ebiten.KeyW: 0x5, ebiten.KeyE: 0x6, ebiten.KeyR: 0xD,
	ebiten.KeyA: 0x7, ebiten.KeyS: 0x8, ebiten.KeyD: 0x9, ebiten.KeyF: 0xE,
	ebiten.KeyZ: 0xA, ebiten.KeyX: 0x0, ebiten.KeyC: 0xB, ebiten.KeyV: 0xF,
}

func (game *Game) Update() error {
	// Input Polling
	for i := range game.chip8.Keypad {
		game.chip8.Keypad[i] = 0
	}
	for ebKey, chip8Val := range keyMap {
		if ebiten.IsKeyPressed(ebKey) {
			game.chip8.Keypad[chip8Val] = 1
		}
	}

	// Toggle Pause
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		game.paused = !game.paused
	}

	// Step forward one cycle (Your exact logic)
	if game.paused && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		game.chip8.Cycle()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		game.nextROM()
	}
	// Normal execution
	if !game.paused {
		// Professional tip: Run 10 cycles here so the game isn't 10x slow
		for i := 0; i < 10; i++ {
			game.chip8.Cycle()
		}
	}

	// Update Timers at 60Hz
	game.chip8.UpdateTimers()

	return nil
}
func (game *Game) loadEmbeddedROM(path string) {
	data, err := romFiles.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read embedded ROM: %v", err)
		return
	}
	game.chip8.LoadFromBytes(data)
}

var currentRomIdx = 0

func (game *Game) nextROM() {
	currentRomIdx = (currentRomIdx + 1) % len(availableROMs)
	game.chip8 = chip8.New() // Reset CPU
	game.loadEmbeddedROM(availableROMs[currentRomIdx])
}

func (game *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	white := color.RGBA{255, 255, 255, 255}
	dim := color.RGBA{100, 100, 100, 255}

	//Draw Game Screen ---
	for y, row := range game.chip8.Video {
		for x := 0; x < 64; x++ {
			if (row>>(63-x))&1 == 1 {
				ebitenutil.DrawRect(screen, float64(x*scale), float64(y*scale), scale, scale, white)
			}
		}
	}

	//Side Panel Setup ---
	debugX := gameWidth + 20
	// Vertical Divider
	ebitenutil.DrawRect(screen, float64(gameWidth), 0, 2, float64(gameHeight), color.RGBA{30, 30, 30, 255})

	status := "RUNNING"
	if game.paused {
		status = "PAUSED"
	}
	ebitenutil.DebugPrintAt(screen, "STATUS: "+status, debugX, 10)

	// Internal Debugger & Instructions ---
	game.chip8.DrawDebugger(screen, debugX, 40, game.paused)

	asm := game.chip8.GetCurrentInstruction()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("OP: %s", asm), debugX, 210)

	// KEYBOARD LEGEND ---
	ebitenutil.DrawRect(screen, float64(debugX), 235, 120, 1, dim) // Divider
	ebitenutil.DebugPrintAt(screen, "CONTROLS", debugX, 245)

	// Category: Hex Pad
	ebitenutil.DebugPrintAt(screen, "HEX PAD:", debugX, 265)
	ebitenutil.DebugPrintAt(screen, "1 2 3 4  ->  1 2 3 C", debugX, 280)
	ebitenutil.DebugPrintAt(screen, "Q W E R  ->  4 5 6 D", debugX, 295)
	ebitenutil.DebugPrintAt(screen, "A S D F  ->  7 8 9 E", debugX, 310)
	ebitenutil.DebugPrintAt(screen, "Z X C V  ->  A 0 B F", debugX, 325)

	// Category: System
	ebitenutil.DebugPrintAt(screen, "SYSTEM:", debugX, 350)
	ebitenutil.DebugPrintAt(screen, "N: NEXT ROM", debugX, 365)
	ebitenutil.DebugPrintAt(screen, "P: PAUSE", debugX, 375)

	if game.paused {
		ebitenutil.DebugPrintAt(screen, "SPACE: STEP cycle", debugX, 385)
	}

	// Active Key Indicator
	ebitenutil.DebugPrintAt(screen, "ACTIVE:", debugX, 405)
	for i := 0; i < 16; i++ {
		if game.chip8.Keypad[i] == 1 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("[%X]", i), debugX+(i*15), 420)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return gameWidth + debugWidth, gameHeight
}

func main() {
	vm := chip8.New()

	game := &Game{
		chip8:  vm,
		paused: false,
	}

	// Load the first ROM by default for WASM
	game.loadEmbeddedROM(availableROMs[0])

	ebiten.SetWindowTitle("CHIP-8 Go Runtime")
	ebiten.SetWindowSize(gameWidth+debugWidth, gameHeight)

	// Start the browser loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
