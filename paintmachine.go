package intcode

import (
	"errors"
	"fmt"
)

type GridPoint struct {
	X, Y int
}
type ShipColor = int
type ShipGrid = map[GridPoint]ShipColor
type Orientation = int

const (
	ORIENTATION_UP Orientation = iota
	ORIENTATION_LEFT
	ORIENTATION_DOWN
	ORIENTATION_RIGHT
)

type PaintMachine struct {
	Position    GridPoint
	Orientation Orientation
	VM          *IntCodeVM
}

func NewPaintMachine(program string) *PaintMachine {
	return &PaintMachine{
		VM: NewVM(program),
	}
}

func (machine *PaintMachine) Paint(ship ShipGrid) error {
	var output int
	var err error

	for {
		machine.VM.Stdin.Write(ship[machine.Position])
		output, err = machine.VM.RunUntilOutput()
		if err != nil {
			return err
		}

		// Paint it
		ship[machine.Position] = output

		// Turn
		output, err = machine.VM.RunUntilOutput()
		if err != nil {
			return err
		}
		switch output {
		case 1: // Turn Left
			machine.Orientation = (machine.Orientation + 1) % 4
		case 0: // Turn Right
			tmp := (machine.Orientation - 1)
			if tmp == -1 {
				tmp = 3
			}
			machine.Orientation = tmp
		default:
			return errors.New(fmt.Sprintf("Unkown direction %d", output))
		}

		switch machine.Orientation {
		case ORIENTATION_LEFT:
			machine.Position.X += 1
		case ORIENTATION_RIGHT:
			machine.Position.X -= 1
		case ORIENTATION_UP:
			machine.Position.Y += 1
		case ORIENTATION_DOWN:
			machine.Position.Y -= 1
		}
	}
}
