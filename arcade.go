package intcode

type ScreenBuffer = [][]int

type Tile = int

const (
	TILE_EMPTY Tile = iota
	TILE_WALL
	TILE_BLOCK
	TILE_HORIZONTAL_PADDLE
	TILE_BALL
)

type TileDrawCommand struct {
	TileId, X, Y int
}

type ArcadeMachine struct {
	Screen ScreenBuffer
	VM     *IntCodeVM
}

func NewArcadeMachine(program string) *ArcadeMachine {
	screen := make([][]int, 100)
	for i, _ := range screen {
		screen[i] = make([]int, 100)
	}
	return &ArcadeMachine{
		Screen: screen,
		VM:     NewVM(program),
	}
}

func (arcade *ArcadeMachine) RunUntilTileDraw() (TileDrawCommand, error) {
	var tileId, x, y int
	var err error
	x, err = arcade.VM.RunUntilOutput()
	if err != nil {
		return TileDrawCommand{}, err
	}
	y, err = arcade.VM.RunUntilOutput()
	if err != nil {
		return TileDrawCommand{}, err
	}
	tileId, err = arcade.VM.RunUntilOutput()
	if err != nil {
		return TileDrawCommand{}, err
	}
	return TileDrawCommand{
		TileId: tileId,
		X:      x,
		Y:      y,
	}, nil
}

func (arcade *ArcadeMachine) Play() error {
	var err error
	var drawCommand TileDrawCommand
	for {
		drawCommand, err = arcade.RunUntilTileDraw()
		if err != nil {
			return err
		}
		arcade.Screen[drawCommand.Y][drawCommand.X] = drawCommand.TileId
	}
}
