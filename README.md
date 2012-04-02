# GeneticGo - genetic algorithm library written in Go

This is a library I'm building while learning to code in Go

## Usage

GeneticGo is compatible with Go 1. Add it to your package repository:

	go get https://github.com/handcraftsman/GeneticGo/tree/master/genetic

Use it in your program:

	import ( "github.com/handcraftsman/GeneticGo/tree/master/genetic" )

## Samples (in order written)

- string_duplication.go - duplicates a string, see [related blog post](http://handcraftsman.wordpress.com/2012/03/27/first-program-in-go-simple-genetic-solver/)

    go run samples/string_duplication.go

- 8queens.go - solves the 8 Queens Puzzle, see [related blog post](http://handcraftsman.wordpress.com/2012/03/30/solving-the-8-queens-puzzle-with-go/)

    go run samples/8queens.go

- tsp.go - travelling salesperson problem solver

	go run samples/tsp.go samples/data/tsp/eil51.tsp

## License		

[MIT License](http://www.opensource.org/licenses/mit-license.php)
