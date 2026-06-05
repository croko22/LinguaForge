package srs

import (
	"math"
	"time"
)

// Quality represents the user's recall performance.
type Quality int

const (
	Again Quality = 1
	Hard  Quality = 2
	Good  Quality = 3
	Easy  Quality = 4
)

// CardInput represents the current state of a review card before review.
type CardInput struct {
	Repetitions int
	EaseFactor  float64
	Interval    int // days
}

// CardOutput represents the computed next review state.
type CardOutput struct {
	Repetitions int
	EaseFactor  float64
	Interval    int // days
	NextReview  time.Time
}

// ComputeNextReview applies the SM-2 algorithm to compute the next review state.
//
// Quality < 3 means failed: reset repetitions, interval = 1 day.
// Quality >= 3 means passed: increase interval based on current state.
func ComputeNextReview(input CardInput, quality Quality, now time.Time) CardOutput {
	ef := input.EaseFactor
	if ef < 1.3 {
		ef = 1.3
	}

	output := CardOutput{}

	if quality < Good {
		// Failed — reset
		output.Repetitions = 0
		output.Interval = 1
	} else {
		// Passed — advance
		output.Repetitions = input.Repetitions + 1

		switch input.Repetitions {
		case 0:
			output.Interval = 1
		case 1:
			output.Interval = 6
		default:
			output.Interval = int(math.Round(float64(input.Interval) * ef))
		}
	}

	// Update ease factor using SM-2 formula (always applied regardless of pass/fail)
	q := float64(quality)
	ef = ef + (0.1 - (5.0-q)*(0.08+(5.0-q)*0.02))
	if ef < 1.3 {
		ef = 1.3
	}
	output.EaseFactor = ef
	output.NextReview = now.AddDate(0, 0, output.Interval)

	return output
}
