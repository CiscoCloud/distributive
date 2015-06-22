package tabular

import (
	"log"
	"math"
	"regexp"
)

// This is a lot of helper functions. To keep things straight, here's a diagram
// of where they're used.
// mean -> sqrDiff -> variance -> stdDev -> normal -> chauvenet -> ProbabalisticSplit
// [fork]          -> chauvenet -> ProbabalisticSplit

// meanInt uses sum to find the mean of a list of ints.
func meanInt(data []int) float64 {
	// sumInt finds the sum of a list of ints. Exciting, I know.
	sumInt := func(data []int) (sum int) {
		for _, i := range data {
			sum += i
		}
		return sum
	}
	return float64(sumInt(data)) / float64(len(data))
}

// meanFloat uses sum to find the mean of a list of float.
func meanFloat(data []float64) float64 {
	// sumFloat finds the sum of a list of floats. Exciting, I know.
	sumFloat := func(data []float64) (sum float64) {
		for _, i := range data {
			sum += i
		}
		return sum
	}
	return sumFloat(data) / float64(len(data))
}

// getSquaredDifferences returns a list of the squared differences between each
// point and the mean. Used in variance and chauvenet->highestVarianceIndex
func getSquaredDifferences(data []int) (sqrDiffs []float64) {
	xbar := meanInt(data)
	for _, xn := range data {
		sqrDiff := math.Pow((xbar - float64(xn)), 2)
		sqrDiffs = append(sqrDiffs, sqrDiff)
	}
	return sqrDiffs
}

// variance is the mean of the squared differences between each observed
// point and the mean (xbar)
func variance(data []int) float64 {
	return meanFloat(getSquaredDifferences(data))
}

// stdDev returns the standard deviation of a data set of integers
func stdDev(data []int) float64 {
	return math.Sqrt(variance(data))
}

// normalDistribution returns the probability that the data point x
// lands where it does, based on the mean (mu) and standard deviation (sigma)
func normalDistribution(x float64, mu float64, sigma float64) float64 {
	coefficient := 1 / (sigma * math.Sqrt(2*math.Pi))
	exponent := -(math.Pow(x-mu, 2) / (2 * math.Pow(sigma, 2)))
	return coefficient * math.Pow(math.E, exponent)
}

type compare func(x float64, y float64) bool

var maxFunc compare = func(x float64, y float64) bool { return x > y }
var minFunc compare = func(x float64, y float64) bool { return x < y }

// extremaIndex finds the index of a value in a list that when compared with
// the any of other data with comparisonFunc will return true. It is most easily
// applicable in finding maxes and mins
func extremaIndex(comparisonFunc compare, data []float64) (index int) {
	if len(data) < 1 {
		return 0
	}
	extrema := data[0]
	extremaIndex := 0
	for i, datum := range data {
		if comparisonFunc(datum, extrema) {
			extrema = datum
			extremaIndex = i
		}
	}
	return extremaIndex
}

// chauvenet simply takes a slice of integers, applies Chauvenet's Criterion,
// and potentially discards a single outlier. Not necessarily, though!
// https://en.wikipedia.org/wiki/Chauvenet%27s_criterion
func chauvenet(data []int) (result []int) {
	// isOutlier applies chauvenet's criterion to determine whether or not
	// x is an outlier
	isOutlier := func(x float64, data []int) bool {
		xbar := meanInt(data)
		sigma := stdDev(data)
		probability := normalDistribution(x, xbar, sigma)
		if probability < float64(1/(2*len(data))) {
			return true
		}
		return false
	}
	// if its an outlier, cut it out. If not, leave it in.
	// find the index with the highest variance
	index := extremaIndex(maxFunc, getSquaredDifferences(data))
	potentialOutlier := float64(data[index])
	// test if that datum is an outlier
	if isOutlier(potentialOutlier, data) {
		return append(data[:index], data[index+1:]...)
	}
	return data
}

// ProbabalisticSplit splits a string based on the regexp that gives the most
// consistent line length (potentially discarding one outlier)
func ProbabalisticSplit(str string) (output Table) {
	// allEqual checks to see if all of the given list of integers are the same
	allEqual := func(ints []int) bool {
		if len(ints) < 1 {
			return true
		}
		sentinel := ints[0]
		for _, i := range ints {
			if i != sentinel {
				return false
			}
		}
		return true
	}
	// notAllOne ensures that the given list of integers are not all == 1
	notAllOne := func(ints []int) bool {
		for _, i := range ints {
			if i != 1 {
				return true
			}
		}
		return false
	}
	// getRowLengths returns row length counts for each table
	getRowLengths := func(tables []Table) (rowLengths [][]int) {
		for _, table := range tables {
			var lengths []int
			for _, row := range table {
				lengths = append(lengths, len(row))
			}
			rowLengths = append(rowLengths, lengths)
		}
		return rowLengths
	}

	// getColumnRegex is the core of the logic. It determines which regex most
	// accurately splits the data into columns by testing the deviation in the
	// row lengths using different regexps.
	// TODO still sucks when the regex doesn't split anything
	// TODO fix: only use that regexp if you find a match at all
	getColumnRegex := func(str string, rowSep *regexp.Regexp) *regexp.Regexp {
		// different column separators to try out
		colSeps := []*regexp.Regexp{
			regexp.MustCompile("\\s+"),    // any whitespace
			regexp.MustCompile("\\s{2,}"), // two+ whitespace (spaces in cols)
			regexp.MustCompile("\\s{4}"),  // exactly four whitespaces
			//regexp.MustCompile("\\t+"),    // tabs
		}
		// separate the data based on the above column regexps
		var tables []Table
		for _, colSep := range colSeps {
			table := SeparateString(rowSep, colSep, str)
			tables = append(tables, table)
		}
		// see if any of them had utterly consistent row lengths
		rowLengths := getRowLengths(tables)
		if len(rowLengths) != len(tables) {
			log.Fatal("Internal error: len(rowLengths) != len(tables)")
		}
		for i, lengths := range rowLengths {
			if allEqual(lengths) && notAllOne(lengths) {
				return colSeps[i]
			}
		}
		// if not, cast out outliers and try again
		for i, lengths := range rowLengths {
			rowLengths[i] = chauvenet(lengths)
		}
		for i, lengths := range rowLengths {
			if allEqual(lengths) && notAllOne(lengths) {
				return colSeps[i]
			}
		}
		// if still not done, just pick the one with the lowest variance
		var variances []float64
		for _, lengths := range rowLengths {
			variances = append(variances, variance(lengths))
		}
		// ensure that index can be found in tables
		minVarianceIndex := extremaIndex(minFunc, variances)
		if len(tables) <= minVarianceIndex {
			log.Fatal("Internal error: minVarianceIndex couldn't be found in tables")
		}
		return colSeps[minVarianceIndex]
	}
	rowSep := regexp.MustCompile("\n+")
	colSep := getColumnRegex(str, rowSep)
	//fmt.Println("THE CHOSEN REGEX: " + colSep.String())
	return SeparateString(rowSep, colSep, str)
}
