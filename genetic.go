package genetic

import (
	"fmt"
	"time"
)

func (solver *Solver) GetBestUsingHillClimbing(getFitness func(string) int,
	display func(string),
	geneSet string,
	maxNumberOfChromosomes, numberOfGenesPerChromosome int,
	bestPossibleFitness int) string {

	solver.initialize(geneSet, numberOfGenesPerChromosome, getFitness)

	roundsSinceLastImprovement := 0
	generationCount := 1

	bestEver := sequenceInfo{genes: ""}
	bestEver.fitness = getFitness(bestEver.genes)
	initialParent := bestEver

	solver.initializePool(generationCount, numberOfGenesPerChromosome, geneSet, initialParent, getFitness)
	solver.initializeStrategies(numberOfGenesPerChromosome, getFitness)

	defer func() {
		solver.quit <- true
		<-solver.nextChromosome
		<-solver.nextGene
		for _, strategy := range solver.strategies {
			select {
			case <-strategy.results:
			default:
			}
		}
	}()

	filteredDisplay := func(item *sequenceInfo) {
		if solver.childFitnessIsBetter(*item, bestEver) {
			display((*item).genes)
			bestEver = *item
		}
	}

	bestEver = solver.pool[0]
	display(bestEver.genes)

	for generationCount <= maxNumberOfChromosomes &&
		roundsSinceLastImprovement < solver.MaxRoundsWithoutImprovement &&
		bestEver.fitness != bestPossibleFitness {

		result := solver.getBestWithInitialParent(getFitness,
			filteredDisplay,
			geneSet,
			generationCount,
			numberOfGenesPerChromosome)

		if solver.childFitnessIsBetter(result, bestEver) {
			roundsSinceLastImprovement = 0
			bestEver = result
			if bestEver.fitness == bestPossibleFitness {
				break
			}
		} else {
			roundsSinceLastImprovement++
			if roundsSinceLastImprovement >= solver.MaxRoundsWithoutImprovement {
				break
			}
		}

		newPool := make([]sequenceInfo, 0, solver.maxPoolSize)
		distinctPool := make(map[string]bool, solver.maxPoolSize)

		improved := false

		generationCount++
		solver.printNewlineIfNecessary()
		initialStrategy := strategyInfo{name: "initial   "}

		for round := 0; round < 100 && !improved; round++ {
			for _, parent := range solver.pool {
				childGenes := parent.genes + <-solver.nextChromosome
				if distinctPool[childGenes] {
					continue
				}
				distinctPool[childGenes] = true

				fitness := getFitness(childGenes)
				child := sequenceInfo{genes: childGenes, fitness: fitness, strategy: initialStrategy}
				if len(newPool) < solver.maxPoolSize {
					newPool = append(newPool, child)
				} else {
					newPool[len(newPool)-1] = child
				}
				insertionSort(newPool, solver.childFitnessIsSameOrBetter, len(newPool)-1)

				if solver.childFitnessIsBetter(child, bestEver) {
					roundsSinceLastImprovement = 0
					solver.printNewlineIfNecessary()
					if solver.PrintStrategyUsage {
						fmt.Print("climb     ")
					}
					display(child.genes)
					bestEver = child
					improved = true
				}
			}
		}

		solver.pool = newPool
		solver.distinctPool = distinctPool
	}

	solver.printNewlineIfNecessary()
	solver.printStrategyUsage()

	return bestEver.genes
}

func (solver *Solver) GetBest(getFitness func(string) int,
	display func(string),
	geneSet string,
	numberOfChromosomes, numberOfGenesPerChromosome int) string {

	solver.initialize(geneSet, numberOfGenesPerChromosome, getFitness)

	defer func() {
		solver.quit <- true
		<-solver.nextChromosome
		<-solver.nextGene
		for _, strategy := range solver.strategies {
			select {
			case <-strategy.results:
			default:
			}
		}
	}()

	initialParent := sequenceInfo{genes: generateParent(solver.nextChromosome, geneSet, numberOfChromosomes, numberOfGenesPerChromosome)}
	initialParent.fitness = getFitness(initialParent.genes)

	solver.initializePool(numberOfChromosomes, numberOfGenesPerChromosome, geneSet, initialParent, getFitness)
	solver.initializeStrategies(numberOfGenesPerChromosome, getFitness)

	best := *new(sequenceInfo)
	displayCaptureBest := func(sequence *sequenceInfo) {
		display((*sequence).genes)
		best = *sequence
	}

	solver.getBestWithInitialParent(getFitness,
		displayCaptureBest,
		geneSet,
		numberOfChromosomes,
		numberOfGenesPerChromosome)

	solver.printNewlineIfNecessary()
	solver.printStrategyUsage()

	return best.genes
}

func (solver *Solver) getBestWithInitialParent(getFitness func(string) int,
	display func(*sequenceInfo),
	geneSet string,
	numberOfChromosomes, numberOfGenesPerChromosome int) sequenceInfo {

	start := time.Now()

	children := make([]sequenceInfo, 1, solver.maxPoolSize)
	children[0] = solver.pool[0]

	distinctChildren := make(map[string]bool, len(solver.pool))
	distinctChildrenFitnesses := populateDistinctPoolFitnessesMap(solver.pool)

	quit := make(chan bool)

	promoteChildrenIfFull := func() {
		if len(children) >= 20 || len(children) >= 10 && time.Since(start).Seconds() > solver.MaxSecondsToRunWithoutImprovement/2 {
			if solver.PrintDiagnosticInfo {
				fmt.Print(">")
				solver.needNewlineBeforeDisplay = true
			}

			solver.poolLock.Lock()
			solver.distinctPoolLock.Lock()
			solver.pool = children
			solver.distinctPool = distinctChildren
			solver.distinctPoolLock.Unlock()
			solver.poolLock.Unlock()

			bestParent := solver.pool[0]
			children = make([]sequenceInfo, 1, solver.maxPoolSize)
			children[0] = bestParent

			distinctChildren = make(map[string]bool, len(children))
			distinctChildren[bestParent.genes] = true

			distinctChildrenFitnesses = make(map[int]bool, len(solver.pool))
			distinctChildrenFitnesses[bestParent.fitness] = true
		}
	}

	updatePools := func(child *sequenceInfo) bool {
		addToChildren := func(item *sequenceInfo) {
			if len(children) < solver.maxPoolSize &&
				(len(distinctChildrenFitnesses) < 4 ||
					(*item).fitness == children[len(children)-1].fitness) {

				children = append(children, *item)

				if solver.PrintDiagnosticInfo {
					fmt.Print(".")
					solver.needNewlineBeforeDisplay = true
				}
				insertionSort(children, solver.childFitnessIsSameOrBetter, len(children)-1)
			} else if solver.childFitnessIsSameOrBetter(*item, children[len(children)-1]) {
				children[len(children)-1] = *item
				insertionSort(children, solver.childFitnessIsSameOrBetter, len(children)-1)
			}

			distinctChildren[(*item).genes] = true
			distinctChildrenFitnesses[(*item).fitness] = true
		}
		addToChildren(child)

		if solver.childFitnessIsBetter(*child, solver.pool[0]) {
			solver.printNewlineIfNecessary()
			if solver.PrintStrategyUsage {
				fmt.Print((*child).strategy.name)
			}
			display(child)

			if solver.pool[0].genes == (*(*child).parent).genes {
				solver.successParentIsBestParentCount++
			}
			solver.numberOfImprovements++

			if solver.PrintDiagnosticInfo {
				fmt.Print("+")
				solver.needNewlineBeforeDisplay = true
			}

			solver.pool[len(solver.pool)-1] = *child
			insertionSort(solver.pool, solver.childFitnessIsSameOrBetter, len(solver.pool)-1)

			if !distinctChildren[(*(*child).parent).genes] {
				addToChildren(child.parent)
			}

			return true
		}

		return false
	}

	updateIfIsImprovement := func(child *sequenceInfo) {
		if solver.shouldAddChild(child, getFitness) {
			if updatePools(child) {
				solver.incrementStrategyUseCount((*child).strategy.index)
				start = time.Now()
			}
		}
	}

	timeout := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(1 * time.Millisecond)
			select {
			case timeout <- true:
			case <-quit:
				quit <- true
			}
		}
		close(timeout)
	}()

	defer func() {
		quit <- true
		for _, child := range children {
			solver.pool = append(solver.pool, child)
		}
	}()

	for {
		// prefer successful strategies
		minStrategySuccess := solver.nextRand(solver.maxStrategySuccess)
		for index := 0; index < len(solver.strategies); index++ {
			if solver.strategies[index].successCount < minStrategySuccess {
				continue
			}
			select {
			case child := <-solver.strategies[index].results:
				updateIfIsImprovement(child)
			case <-timeout:
				if time.Since(start).Seconds() >= solver.MaxSecondsToRunWithoutImprovement {
					return solver.pool[0]
				}
				promoteChildrenIfFull()
			}
		}
	}
	return solver.pool[0]
}

func (solver *Solver) createFitnessComparisonFunctions() {
	if solver.LowerFitnessesAreBetter {
		solver.childFitnessIsBetter = func(child, other sequenceInfo) bool {
			return child.fitness < other.fitness
		}

		solver.childFitnessIsSameOrBetter = func(child, other sequenceInfo) bool {
			return child.fitness <= other.fitness
		}
	} else {
		solver.childFitnessIsBetter = func(child, other sequenceInfo) bool {
			return child.fitness > other.fitness
		}

		solver.childFitnessIsSameOrBetter = func(child, other sequenceInfo) bool {
			return child.fitness >= other.fitness
		}
	}
}

func (solver *Solver) ensureMaxSecondsToRunIsValid() {
	if solver.MaxSecondsToRunWithoutImprovement == 0 {
		solver.MaxSecondsToRunWithoutImprovement = 20
		fmt.Printf("\tSolver will run at most %v second(s) without improvement.\n", solver.MaxSecondsToRunWithoutImprovement)
	}
}

func (solver *Solver) incrementStrategyUseCount(strategyIndex int) {
	solver.strategies[strategyIndex].successCount++
	if solver.strategies[strategyIndex].successCount > solver.maxStrategySuccess {
		solver.maxStrategySuccess = solver.strategies[strategyIndex].successCount
	}
}

func (solver *Solver) initialize(geneSet string, numberOfGenesPerChromosome int, getFitness func(string) int) {
	if solver.RandSeed == 0 {
		solver.RandSeed = time.Now().UnixNano()
	}
	if solver.MaxRoundsWithoutImprovement == 0 {
		solver.MaxRoundsWithoutImprovement = 2
	}
	solver.random = createRandomNumberGenerator(solver.RandSeed)
	solver.ensureMaxSecondsToRunIsValid()
	solver.createFitnessComparisonFunctions()
	solver.initializeChannels(geneSet, numberOfGenesPerChromosome)
	solver.needNewlineBeforeDisplay = false
}

func (solver *Solver) initializeChannels(geneSet string, numberOfGenesPerChromosome int) {
	solver.quit = make(chan bool)
	solver.nextGene = make(chan string, 1+numberOfGenesPerChromosome)
	go generateGene(solver.nextGene, geneSet, solver.quit, solver.RandSeed)

	solver.nextChromosome = make(chan string, 1)
	go generateChromosome(solver.nextChromosome, solver.nextGene, geneSet, numberOfGenesPerChromosome, solver.quit)
}

func (solver *Solver) initializePool(numberOfChromosomes, numberOfGenesPerChromosome int, geneSet string, initialParent sequenceInfo, getFitness func(string) int) {
	solver.maxPoolSize = max(len(geneSet), 3*numberOfChromosomes*numberOfGenesPerChromosome)
	solver.pool = make([]sequenceInfo, solver.maxPoolSize, solver.maxPoolSize)
	solver.pool[0] = initialParent
	solver.distinctPool = populatePool(solver.pool, solver.nextChromosome, geneSet, numberOfChromosomes, numberOfGenesPerChromosome, solver.childFitnessIsBetter, getFitness)

	solver.numberOfImprovements = 1
	solver.randomParent = make(chan *sequenceInfo, 10)
	go func() {
		for {
			select {
			case <-solver.quit:
				solver.quit <- true
				return
			default:
				useBestParent := solver.random.Intn(solver.numberOfImprovements) <= solver.successParentIsBestParentCount
				if useBestParent {
					parent := solver.pool[0]
					solver.randomParent <- &parent
				}
				parent := solver.pool[solver.random.Intn(len(solver.pool))]
				solver.randomParent <- &parent
			}
		}
	}()
}

func (solver *Solver) nextRand(limit int) int {
	return solver.random.Intn(limit)
}

func (solver *Solver) printNewlineIfNecessary() {
	if solver.needNewlineBeforeDisplay {
		solver.needNewlineBeforeDisplay = false
		fmt.Println()
	}
}

func (solver *Solver) printStrategyUsage() {
	if !solver.PrintStrategyUsage {
		return
	}

	fmt.Println("\nstrategy usage:")
	for _, strategy := range solver.strategies {
		fmt.Println(
			strategy.name, "\t",
			strategy.successCount, "\t",
			100.0*strategy.successCount/solver.numberOfImprovements, "%")
	}
	fmt.Println()

	fmt.Println("\nNew champions were children of the reigning champion",
		100*solver.successParentIsBestParentCount/solver.numberOfImprovements,
		"% of the time.")
}

func (solver *Solver) shouldAddChild(child *sequenceInfo, getFitness func(string) int) bool {
	if solver.inPool((*child).genes) {
		return false
	}

	(*child).fitness = getFitness((*child).genes)
	if !solver.childFitnessIsSameOrBetter(*child, solver.pool[len(solver.pool)-1]) {
		return false
	}

	if (*child).fitness == solver.pool[len(solver.pool)-1].fitness {
		if len(solver.pool) < solver.maxPoolSize {
			solver.pool = append(solver.pool, *child)
		} else {
			solver.pool[len(solver.pool)-1] = *child
			insertionSort(solver.pool, solver.childFitnessIsSameOrBetter, len(solver.pool)-1)
		}

		return false
	}

	return true
}

func populateDistinctPoolFitnessesMap(pool []sequenceInfo) map[int]bool {
	distinctChildrenFitnesses := make(map[int]bool, len(pool))
	distinctChildrenFitnesses[pool[0].fitness] = true
	distinctChildrenFitnesses[pool[len(pool)/2].fitness] = true
	distinctChildrenFitnesses[pool[len(pool)-1].fitness] = true
	return distinctChildrenFitnesses
}
