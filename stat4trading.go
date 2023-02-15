package stat4trading

import "errors"

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
