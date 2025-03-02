package flags

import "flag"

type ApplicationArguments struct {
	AmountBenchmarkingSamples, ParameterSet int
}

func GetApplicationArguments() ApplicationArguments {
	// Creating struct with empty arguments
	arguments := ApplicationArguments{}

	// Getting arguments from flags
	flag.IntVar(&arguments.AmountBenchmarkingSamples, "b", 0,
		"Decides if the implementation should be benchmarked, and the amount of samples")
	flag.IntVar(&arguments.ParameterSet, "p", 0,
		"Decides what parameter set should be used")

	// Parsing flags
	flag.Parse()

	return arguments
}
