package intcode

type Amplifier struct {
	PhaseSetting int
	VM           *IntCodeVM
}

func NewAmplifier(phaseSetting int, program string) *Amplifier {
	return &Amplifier{
		PhaseSetting: phaseSetting,
		VM:           NewVM(program),
	}
}

func (amp *Amplifier) Connect(other *Amplifier) {
	amp.VM.Stdout = other.VM.Stdin
}

func (amp *Amplifier) Init() {
	amp.VM.Stdin.Write(amp.PhaseSetting)
}

func OptimizeAmplifierPhaseSettings(program string) ([]int, int, error) {
	best := -1
	var bestp []int
	for permutation := range GeneratePermutations([]int{0, 1, 2, 3, 4}) {
		amps := make([]*Amplifier, 0, len(permutation))
		// Make an amps for each setting
		for _, setting := range permutation {
			amps = append(amps, NewAmplifier(setting, program))
		}
		// connect 'em up
		for i, amp := range amps[:len(amps)-1] {
			amp.Connect(amps[i+1])
		}
		// init
		for _, amp := range amps {
			amp.Init()
		}
		// Initial Input
		amps[0].VM.Stdin.Write(0)
		// Run
		var err error
		for _, amp := range amps {
			err = amp.VM.RunUntilHalt()
			if err != nil {
				return nil, 0, err
			}
		}
		result, err := amps[len(amps)-1].VM.Stdout.Read()
		//log.Printf("%v\n", amps[len(amps)-1].VM.Stdout.(*IntCodePipe).Buffer)
		if result > best {
			best = result
			bestp = permutation
		}
	}
	return bestp, best, nil
}

func OptimizeLoopedAmplifierPhaseSettings(program string) ([]int, int, error) {
	best := -1
	var bestp []int
	for permutation := range GeneratePermutations([]int{5, 6, 7, 8, 9}) {
		amps := make([]*Amplifier, 0, len(permutation))
		// Make an amps for each setting
		for _, setting := range permutation {
			amps = append(amps, NewAmplifier(setting, program))
		}
		// connect 'em up
		for i, amp := range amps[:len(amps)-1] {
			amp.Connect(amps[i+1])
		}
		amps[0].VM.Stdin = amps[len(amps)-1].VM.Stdout
		// init
		for _, amp := range amps {
			amp.Init()
		}
		// Initial Input
		amps[0].VM.Stdin.Write(0)
		// Run
		var err error
		halts := 0
		for halts < len(amps) {
			for _, amp := range amps {
				err = amp.VM.RunUntilError()
				switch err {
				case ErrHalt:
					halts += 1
				case ReadError: //
					break
				default:
					return nil, 0, err
				}
			}
		}
		result, err := amps[len(amps)-1].VM.Stdout.Read()
		if result > best {
			best = result
			bestp = permutation
		}
	}
	return bestp, best, nil
}
