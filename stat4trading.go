package stat4trading

import "errors"

type Numeric interface {
	int64 | float64 | int32 | float32 | int
}

// SMA - SimpleMovingAverage
// expectedOutputDataLength is optional parameter, if you don't want to control the length, set expectedOutputDataLength = -1
func SMA(inputData []float64, windowWidth int, expectedOutputDataLength int) ([]float64, error) {
	outputDataLength := len(inputData) - windowWidth + 1

	if expectedOutputDataLength > 0 && expectedOutputDataLength != outputDataLength {
		return nil, errors.New("incorrectly calculated expected output data length")
	}

	if outputDataLength <= 0 {
		return nil, errors.New("not enough data to calculate SMA of specified window width, increase data set or reduce window width")
	}

	processedData := make([]float64, outputDataLength)

	for i := 0; i < outputDataLength; i++ {
		windowStart := i
		windowEnd := i + windowWidth - 1
		sum := 0.0

		for j := windowStart; j <= windowEnd; j++ {
			sum += inputData[j]
		}

		processedData[i] = sum / float64(windowWidth)
	}

	return processedData, nil
}

// WMA - WeightedMovingAverage
// expectedOutputDataLength is optional parameter, if you don't want to control the length, set expectedOutputDataLength = -1
func WMA(inputData []float64, windowWidth int, expectedOutputDataLength int) ([]float64, error) {
	outputDataLength := len(inputData) - windowWidth + 1

	if expectedOutputDataLength > 0 && expectedOutputDataLength != outputDataLength {
		return nil, errors.New("incorrectly calculated expected output data length")
	}

	if outputDataLength <= 0 {
		return nil, errors.New("not enough data to calculate WMA of specified window width, increase data set or reduce window width")
	}

	processedData := make([]float64, outputDataLength)

	// https://ru.wikipedia.org/wiki/%D0%A1%D0%BA%D0%BE%D0%BB%D1%8C%D0%B7%D1%8F%D1%89%D0%B0%D1%8F_%D1%81%D1%80%D0%B5%D0%B4%D0%BD%D1%8F%D1%8F
	denominator := float64(windowWidth * (windowWidth + 1) / 2)

	for i := 0; i < outputDataLength; i++ {
		windowStart := i
		windowEnd := i + windowWidth - 1
		sum := 0.0

		for j := windowStart; j <= windowEnd; j++ {
			linearlyIncreasingFactor := float64(j - windowStart + 1) // [1, 2, 3, ... windowWidth]
			sum += inputData[j] * linearlyIncreasingFactor
		}

		processedData[i] = sum / denominator
	}

	return processedData, nil
}

// EMA - ExponentialMovingAverage
// expectedOutputDataLength is optional parameter, if you don't want to control the length, set expectedOutputDataLength = -1
func EMA(inputData []float64, windowWidth int, expectedOutputDataLength int) ([]float64, error) {
	// Strictly said, data length after EMA should be different in comparing to SMA and WMA but we do the same for unification
	outputDataLength := len(inputData) - windowWidth + 1

	if expectedOutputDataLength > 0 && expectedOutputDataLength != outputDataLength {
		return nil, errors.New("incorrectly calculated expected output data length")
	}

	if outputDataLength <= 0 {
		return nil, errors.New("not enough data to calculate EMA of specified window width, increase data set or reduce window width")
	}

	ema := make([]float64, len(inputData))
	ema[0] = inputData[0]
	alpha := float64(2 / (1 + windowWidth))

	for i := 1; i < len(inputData); i++ {
		ema[i] = alpha*inputData[i] + (1-alpha)*ema[i-1]
	}

	result := ema[windowWidth-1:]

	if len(result) != outputDataLength {
		return nil, errors.New("self-control failed: incorrectly calculated expected output data length")
	}

	return result, nil
}

func Subtract(initialData []float64, deductibleData []float64) ([]float64, error) {
	if len(initialData) != len(deductibleData) {
		return nil, errors.New("input data should be the same length!")
	}

	result := make([]float64, len(initialData))

	for i := 0; i < len(initialData); i++ {
		result[i] = initialData[i] - deductibleData[i]
	}

	return result, nil
}

func FindIntersections(referenceGraph []float64, investigatedGraph []float64) ([]string, error) {
	if len(referenceGraph) != len(investigatedGraph) {
		return nil, errors.New("input data should be the same length!")
	}

	result := make([]string, len(referenceGraph))
	result[0] = ""

	for i := 1; i < len(referenceGraph); i++ {
		if referenceGraph[i-1] > investigatedGraph[i-1] && referenceGraph[i] < investigatedGraph[i] {
			result[i] = "BOTTOM-TO-TOP"
		} else if referenceGraph[i-1] < investigatedGraph[i-1] && referenceGraph[i] > investigatedGraph[i] {
			result[i] = "TOP-TO-BOTTOM"
		} else {
			result[i] = ""
		}
	}

	return result, nil
}

func FindMax[N Numeric](dataSet []N) (N, error) {
	if len(dataSet) == 0 {
		return 0, errors.New("Input data set cannot be empty!")
	}

	maxValue := dataSet[0]

	for i := 1; i < len(dataSet); i++ {
		if dataSet[i] > maxValue {
			maxValue = dataSet[i]
		}
	}

	return maxValue, nil
}

func FindMin[N Numeric](dataSet []N) (N, error) {
	if len(dataSet) == 0 {
		return 0, errors.New("Input data set cannot be empty!")
	}

	minValue := dataSet[0]

	for i := 1; i < len(dataSet); i++ {
		if dataSet[i] < minValue {
			minValue = dataSet[i]
		}
	}

	return minValue, nil
}
