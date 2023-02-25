package stat4trading

import (
	"errors"
	"math"
)
import "fmt"

type Numeric interface {
	int64 | float64 | int32 | float32 | int
}

type PointCoordinates struct {
	X float64
	Y float64
}

type LineDefinedByTwoPoints struct {
	PointA PointCoordinates
	PointB PointCoordinates
}

// CalculateOutputDataLengthAfterMA
// Calculates output data length after applying Moving Average window with width = windowWidth
// to incoming data set with length = inputDataLength
func CalculateOutputDataLengthAfterMA(inputDataLength, windowWidth int) int {
	return inputDataLength - windowWidth + 1
}

// SMA - SimpleMovingAverage
// expectedOutputDataLength is the required parameter for self-control.
// It should be known BEFORE doing smoothing, and if it is calculated incorrectly you can't handle obtained result in a right way.
func SMA(inputData []float64, windowWidth int, expectedOutputDataLength int) ([]float64, error) {
	outputDataLength := CalculateOutputDataLengthAfterMA(len(inputData), windowWidth)

	if outputDataLength <= 0 {
		return nil, errors.New("stat4trading::SMA: not enough data to calculate SMA of specified window width, increase data set or reduce window width")
	}

	if expectedOutputDataLength != outputDataLength {
		return nil, errors.New("stat4trading::SMA: incorrectly calculated expected output data length")
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
// expectedOutputDataLength is the required parameter for self-control.
// It should be known BEFORE doing smoothing, and if it is calculated incorrectly you can't handle obtained result in a right way.
func WMA(inputData []float64, windowWidth int, expectedOutputDataLength int) ([]float64, error) {
	outputDataLength := CalculateOutputDataLengthAfterMA(len(inputData), windowWidth)

	if outputDataLength <= 0 {
		return nil, errors.New("stat4trading::WMA: not enough data to calculate WMA of specified window width, increase data set or reduce window width")
	}

	if expectedOutputDataLength != outputDataLength {
		return nil, errors.New("stat4trading::WMA: incorrectly calculated expected output data length")
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
// expectedOutputDataLength is the required parameter for self-control.
// It should be known BEFORE doing smoothing, and if it is calculated incorrectly you can't handle obtained result in a right way.
// WARNING: Strictly said, when calculating EMA, we should CUT OFF FIRST windowWidth elements before return the result - in contrast to calculating SMA / WMA,
// But we cut off first windowWidth-1 elements in order to unify the result and make it the SAME LENGTH as the SMA and WMA result.
func EMA(inputData []float64, windowWidth int, expectedOutputDataLength int) ([]float64, error) {
	// Strictly said, data length after EMA should be different in comparing to SMA and WMA, but we do the same for unification
	outputDataLength := CalculateOutputDataLengthAfterMA(len(inputData), windowWidth)

	if outputDataLength <= 0 {
		return nil, errors.New("stat4trading::EMA: not enough data to calculate EMA of specified window width, increase data set or reduce window width")
	}

	if expectedOutputDataLength != outputDataLength {
		return nil, errors.New("stat4trading::EMA: incorrectly calculated expected output data length")
	}

	ema := make([]float64, len(inputData))
	ema[0] = inputData[0]
	alpha := float64(2) / float64(1 + windowWidth)

	for i := 1; i < len(inputData); i++ {
		ema[i] = alpha*inputData[i] + (1-alpha)*ema[i-1]
	}

	result := ema[windowWidth-1:]

	if len(result) != outputDataLength {
		return nil, errors.New("stat4trading::EMA: self-control failed: incorrectly calculated expected output data length")
	}

	return result, nil
}

func Subtract(initialData []float64, deductibleData []float64) ([]float64, error) {
	if len(initialData) != len(deductibleData) {
		return nil, errors.New("stat4trading::Subtract: both input data sets should be the same length")
	}

	result := make([]float64, len(initialData))

	for i := 0; i < len(initialData); i++ {
		result[i] = initialData[i] - deductibleData[i]
	}

	return result, nil
}

func FindIntersectionDirections(referenceGraph []float64, investigatedGraph []float64) ([]string, error) {
	if len(referenceGraph) != len(investigatedGraph) {
		return nil, errors.New("stat4trading::FindIntersectionDirections: both input data sets should be the same length")
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

// FindIntersectionPointOfTwoSegments - tries to solve a system of two linear equations and returns 3 parameters:
// 1. Coordinates of intersection point if they are exist
// 2. Boolean indicating if solution exists, or it does not
//		for example, solution does not exist if:
//		a) two lines are parallel,
//		b) if intersection point exists, but it is outside of the common projection
// 3. Error in other abnormal situation.
// PLEASE NOTE: This function searches intersection of SEGMENTS, NOT LINES!!!
func FindIntersectionPointOfTwoSegments(lineA LineDefinedByTwoPoints, lineB LineDefinedByTwoPoints) (PointCoordinates, bool, error) {
	// lineA: y = kx+b
	// lineB: y = mx+c

	deltaXA := lineA.PointB.X - lineA.PointA.X
	deltaXB := lineB.PointB.X - lineB.PointA.X

	if deltaXA <= 1e-9 || deltaXB <= 1e-9 {
		return PointCoordinates{}, false, errors.New("stat4trading::FindIntersectionPointOfTwoLines error: deltaX = x2-x1 = 0, while it should not be so. There is an error in input data")
	}

	k := (lineA.PointB.Y - lineA.PointA.Y) / deltaXA
	b := lineA.PointA.Y - k*lineA.PointA.X

	m := (lineB.PointB.Y - lineB.PointA.Y) / deltaXB
	c := lineB.PointA.Y - m*lineB.PointA.X

	if math.Abs(k-m) <= 1e-9 {
		// Solution DOES NOT exist, but the situation IS REGULAR!
		return PointCoordinates{}, false, nil
	}

	x := (c - b) / (k - m)

	y1 := k*x + b
	y2 := m*x + c

	// This case should never happen, and here just for self-control:
	if y1-y2 > 1e-9 {
		fmt.Printf("Y1 = %.10f\nY2 = %.10f\n", y1, y2)
		panic("Self-control failed: Error in linear equation solving logic!")
	}

	// We found that LINES are intersect, now let's check if SEGMENTS are intersect!
	commonProjectionStartX, _ := FindMax([]float64{lineA.PointA.X, lineB.PointA.X})
	commonProjectionEndX, _ := FindMin([]float64{lineA.PointB.X, lineB.PointB.X})

	// Common projection is FROM commonProjectionStartX TO commonProjectionEndX, not vise versa!
	if commonProjectionStartX <= x && x <= commonProjectionEndX {
		return PointCoordinates{X: x, Y: y1}, true, nil
	}

	return PointCoordinates{}, false, nil
}

func FindMax[N Numeric](dataSet []N) (N, error) {
	if len(dataSet) == 0 {
		return 0, errors.New("stat4trading::FindMax: Input data set cannot be empty!")
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
		return 0, errors.New("stat4trading::FindMin: Input data set cannot be empty!")
	}

	minValue := dataSet[0]

	for i := 1; i < len(dataSet); i++ {
		if dataSet[i] < minValue {
			minValue = dataSet[i]
		}
	}

	return minValue, nil
}
