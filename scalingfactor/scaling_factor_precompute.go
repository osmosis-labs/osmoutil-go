package scalingfactor

import (
	"math"

	sdkmath "cosmossdk.io/math"
)

var (
	exponentToScalingFactorMap    = map[int]float64{}
	exponentToScalingFactorDecMap = map[int]sdkmath.LegacyDec{}
)

func init() {
	for i := 0; i < 36; i++ {
		scalingFactor := math.Pow(10, float64(i))

		exponentToScalingFactorMap[i] = scalingFactor
		exponentToScalingFactorDecMap[i] = sdkmath.LegacyNewDec(int64(scalingFactor))
	}
}

// GetScalingFactor returns a float64 scaling factor for the given exponent
func GetScalingFactor(exponent int) float64 {
	return exponentToScalingFactorMap[exponent]
}

// GetScalingFactorDec returns a LegacyDec scaling factor for the given exponent
func GetScalingFactorDec(exponent int) sdkmath.LegacyDec {
	return exponentToScalingFactorDecMap[exponent]
}
