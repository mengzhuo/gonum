// Copyright ©2014 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package distuv

import (
	"math"
	"math/rand/v2"
	"sort"
	"testing"
)

func TestExponentialProb(t *testing.T) {
	t.Parallel()
	pts := []univariateProbPoint{
		{
			loc:     0,
			prob:    1,
			cumProb: 0,
			logProb: 0,
		},
		{
			loc:     -1,
			prob:    0,
			cumProb: 0,
			logProb: math.Inf(-1),
		},
		{
			loc:     1,
			prob:    1 / (math.E),
			cumProb: 0.6321205588285576784044762298385391325541888689682321654921631983025385042551001966428527256540803563,
			logProb: -1,
		},
		{
			loc:     20,
			prob:    math.Exp(-20),
			cumProb: 0.999999997938846377561442172034059619844179023624192724400896307027755338370835976215440646720089072,
			logProb: -20,
		},
	}
	testDistributionProbs(t, Exponential{Rate: 1}, "Exponential", pts)
}

func TestExponentialFitPrior(t *testing.T) {
	t.Parallel()
	testConjugateUpdate(t, func() ConjugateUpdater { return &Exponential{Rate: 13.7, Src: rand.NewPCG(1, 1)} })
}

func TestExponential(t *testing.T) {
	t.Parallel()
	src := rand.New(rand.NewPCG(1, 1))
	for i, dist := range []Exponential{
		{Rate: 3, Src: src},
		{Rate: 1.5, Src: src},
		{Rate: 0.9, Src: src},
	} {
		testExponential(t, dist, i)
	}
}

func testExponential(t *testing.T, dist Exponential, i int) {
	const (
		tol  = 1e-2
		n    = 3e6
		bins = 50
	)
	x := make([]float64, n)
	generateSamples(x, dist)
	sort.Float64s(x)

	checkMean(t, i, x, dist, tol)
	checkVarAndStd(t, i, x, dist, tol)
	checkEntropy(t, i, x, dist, tol)
	checkExKurtosis(t, i, x, dist, 3e-2)
	checkSkewness(t, i, x, dist, tol)
	checkMedian(t, i, x, dist, tol)
	checkQuantileCDFSurvival(t, i, x, dist, tol)
	checkProbContinuous(t, i, x, 0, math.Inf(1), dist, 1e-10)
	checkProbQuantContinuous(t, i, x, dist, tol)

	if dist.Mode() != 0 {
		t.Errorf("Mode is not 0. Got %v", dist.Mode())
	}

	if dist.NumParameters() != 1 {
		t.Errorf("NumParameters is not 1. Got %v", dist.NumParameters())
	}

	if dist.NumSuffStat() != 1 {
		t.Errorf("NumSuffStat is not 1. Got %v", dist.NumSuffStat())
	}

	scoreInput := dist.ScoreInput(-0.0001)
	if scoreInput != 0 {
		t.Errorf("ScoreInput is not 0 for a negative argument. Got %v", scoreInput)
	}
	scoreInput = dist.ScoreInput(0)
	if !math.IsNaN(scoreInput) {
		t.Errorf("ScoreInput is not NaN at 0. Got %v", scoreInput)
	}
	scoreInput = dist.ScoreInput(1)
	if scoreInput != -dist.Rate {
		t.Errorf("ScoreInput mismatch for a positive argument. Got %v, want %g", scoreInput, dist.Rate)
	}

	deriv := make([]float64, 1)
	dist.Score(deriv, -0.0001)
	if deriv[0] != 0 {
		t.Errorf("Score is not 0 for a negative argument. Got %v", deriv[0])
	}
	dist.Score(deriv, 0)
	if !math.IsNaN(deriv[0]) {
		t.Errorf("Score is not NaN at 0. Got %v", deriv[0])
	}

	if !panics(func() { dist.Quantile(-0.0001) }) {
		t.Errorf("Expected panic with negative argument to Quantile")
	}
	if !panics(func() { dist.Quantile(1.0001) }) {
		t.Errorf("Expected panic with argument to Quantile above 1")
	}
}

func TestExponentialScore(t *testing.T) {
	t.Parallel()
	for _, test := range []*Exponential{
		{
			Rate: 1,
		},
		{
			Rate: 0.35,
		},
		{
			Rate: 4.6,
		},
	} {
		testDerivParam(t, test)
	}
}

func TestExponentialFitPanic(t *testing.T) {
	t.Parallel()
	e := Exponential{Rate: 2}
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("unexpected panic for Fit call: %v", r)
		}
	}()
	e.Fit(make([]float64, 10), nil)
}

func TestExponentialCDFSmallArgument(t *testing.T) {
	t.Parallel()
	e := Exponential{Rate: 1}
	x := 1e-17
	p := e.CDF(x)
	if math.Abs(p-x) > 1e-20 {
		t.Errorf("Wrong CDF value for small argument. Got: %v, want: %g", p, x)
	}
}
