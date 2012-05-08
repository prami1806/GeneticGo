package genetic

type Solver struct {
	MaxSecondsToRunWithoutImprovement float64
	MaxRoundsWithoutImprovement       int
	LowerFitnessesAreBetter           bool
	PrintStrategyUsage                bool
	PrintDiagnosticInfo               bool

	childFitnessIsBetter, childFitnessIsSameOrBetter func(child, other sequenceInfo) bool
	printDiagnostic                                  func(string)

	quit                     chan bool
	nextGene, nextChromosome chan string
	randomParent             chan *sequenceInfo

	strategies                     []strategyInfo
	maxStrategySuccess             int
	numberOfImprovements           int
	successParentIsBestParentCount int

	needNewlineBeforeDisplay bool

	pool        *pool
	maxPoolSize int

	random randomSource

	initialParent  string
	isHillClimbing bool
	geneSet        string
}

type sequenceInfo struct {
	genes    string
	fitness  int
	strategy strategyInfo
	parent   *sequenceInfo
}

type strategyInfo struct {
	name         string
	start        func(strategyIndex int)
	successCount int
	results      chan *sequenceInfo
	index        int
}

type randomSource interface {
	Intn(exclusiveMax int) int
}
