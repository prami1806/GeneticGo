package genetic

type Solver struct {
	RandSeed                          int64
	MaxSecondsToRunWithoutImprovement float64
	LowerFitnessesAreBetter           bool
	PrintStrategyUsage                bool
	PrintDiagnosticInfo               bool

	childFitnessIsBetter, childFitnessIsSameOrBetter func(child, other sequenceInfo) bool

	quit                     bool
	nextGene, nextChromosome chan string

	strategies         []strategyInfo
	strategySuccessSum int

	needNewlineBeforeDisplay bool

	maxPoolSize  int
	pool         []sequenceInfo
	distinctPool map[string]bool
}

type sequenceInfo struct {
	genes   string
	fitness int
}

type strategyInfo struct {
	name                         string
	function                     func(parentA, parentB, geneSet string, numberOfGenesPerChromosome int, nextGene chan string, useBestParent bool) string
	count                        int
	incrementParentSuccessCounts bool
}
