package srs

import (
	"math"
	"testing"
	"time"
)

func roundFloat(f float64) float64 {
	return math.Round(f*100) / 100
}

func TestComputeNextReview_NewCard_Good(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	result := ComputeNextReview(CardInput{
		Repetitions: 0,
		EaseFactor:  2.5,
		Interval:    0,
	}, Good, now)

	if result.Repetitions != 1 {
		t.Errorf("expected repetitions=1, got %d", result.Repetitions)
	}
	if result.Interval != 1 {
		t.Errorf("expected interval=1, got %d", result.Interval)
	}
	expectedEF := 2.36
	if roundFloat(result.EaseFactor) != expectedEF {
		t.Errorf("expected ease_factor=%.2f, got %.2f", expectedEF, result.EaseFactor)
	}
	expectedNext := now.AddDate(0, 0, 1)
	if !result.NextReview.Equal(expectedNext) {
		t.Errorf("expected next_review=%v, got %v", expectedNext, result.NextReview)
	}
}

func TestComputeNextReview_NewCard_Easy(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	result := ComputeNextReview(CardInput{
		Repetitions: 0,
		EaseFactor:  2.5,
		Interval:    0,
	}, Easy, now)

	if result.Repetitions != 1 {
		t.Errorf("expected repetitions=1, got %d", result.Repetitions)
	}
	if result.Interval != 1 {
		t.Errorf("expected interval=1, got %d", result.Interval)
	}
	expectedEF := 2.5
	if roundFloat(result.EaseFactor) != expectedEF {
		t.Errorf("expected ease_factor=%.2f, got %.2f", expectedEF, result.EaseFactor)
	}
}

func TestComputeNextReview_FailedReview(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	result := ComputeNextReview(CardInput{
		Repetitions: 5,
		EaseFactor:  2.5,
		Interval:    20,
	}, Again, now)

	if result.Repetitions != 0 {
		t.Errorf("expected repetitions=0 after fail, got %d", result.Repetitions)
	}
	if result.Interval != 1 {
		t.Errorf("expected interval=1 after fail, got %d", result.Interval)
	}
	expectedEF := 1.96
	if roundFloat(result.EaseFactor) != expectedEF {
		t.Errorf("expected ease_factor=%.2f, got %.2f", expectedEF, result.EaseFactor)
	}
	expectedNext := now.AddDate(0, 0, 1)
	if !result.NextReview.Equal(expectedNext) {
		t.Errorf("expected next_review=%v, got %v", expectedNext, result.NextReview)
	}
}

func TestComputeNextReview_SecondReview_Good(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	result := ComputeNextReview(CardInput{
		Repetitions: 1,
		EaseFactor:  2.36,
		Interval:    1,
	}, Good, now)

	if result.Repetitions != 2 {
		t.Errorf("expected repetitions=2, got %d", result.Repetitions)
	}
	if result.Interval != 6 {
		t.Errorf("expected interval=6 for second pass, got %d", result.Interval)
	}
	expectedEF := 2.22
	if roundFloat(result.EaseFactor) != expectedEF {
		t.Errorf("expected ease_factor=%.2f, got %.2f", expectedEF, result.EaseFactor)
	}
	expectedNext := now.AddDate(0, 0, 6)
	if !result.NextReview.Equal(expectedNext) {
		t.Errorf("expected next_review=%v, got %v", expectedNext, result.NextReview)
	}
}

func TestComputeNextReview_ThirdReview_Good(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	result := ComputeNextReview(CardInput{
		Repetitions: 2,
		EaseFactor:  2.22,
		Interval:    6,
	}, Good, now)

	if result.Repetitions != 3 {
		t.Errorf("expected repetitions=3, got %d", result.Repetitions)
	}
	expectedInterval := int(math.Round(6 * 2.22))
	if result.Interval != expectedInterval {
		t.Errorf("expected interval=%d (6 * 2.22), got %d", expectedInterval, result.Interval)
	}
	expectedEF := 2.08
	if roundFloat(result.EaseFactor) != expectedEF {
		t.Errorf("expected ease_factor=%.2f, got %.2f", expectedEF, result.EaseFactor)
	}
	expectedNext := now.AddDate(0, 0, expectedInterval)
	if !result.NextReview.Equal(expectedNext) {
		t.Errorf("expected next_review=%v, got %v", expectedNext, result.NextReview)
	}
}

func TestComputeNextReview_EaseFactorFloor(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	// Input EF is already at minimum; applying Good formula would push below floor
	result := ComputeNextReview(CardInput{
		Repetitions: 0,
		EaseFactor:  1.3,
		Interval:    0,
	}, Good, now)

	if result.Repetitions != 1 {
		t.Errorf("expected repetitions=1, got %d", result.Repetitions)
	}
	// EF after formula: 1.3 + (0.1 - 0.24) = 1.16 → clamped to 1.3
	if result.EaseFactor < 1.3 {
		t.Errorf("expected ease_factor >= 1.3, got %f", result.EaseFactor)
	}
	if roundFloat(result.EaseFactor) != 1.3 {
		t.Errorf("expected ease_factor=1.30, got %.2f", result.EaseFactor)
	}
}

func TestComputeNextReview_HardReview(t *testing.T) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)

	result := ComputeNextReview(CardInput{
		Repetitions: 3,
		EaseFactor:  2.5,
		Interval:    20,
	}, Hard, now)

	// Hard (quality=2) is < 3, so it's a failed review
	if result.Repetitions != 0 {
		t.Errorf("expected repetitions=0 after hard (fail), got %d", result.Repetitions)
	}
	if result.Interval != 1 {
		t.Errorf("expected interval=1 after fail, got %d", result.Interval)
	}
	// EF change = 0.1 - 3*(0.08+3*0.02) = 0.1 - 3*0.14 = 0.1 - 0.42 = -0.32
	// 2.5 - 0.32 = 2.18
	expectedEF := 2.18
	if roundFloat(result.EaseFactor) != expectedEF {
		t.Errorf("expected ease_factor=%.2f, got %.2f", expectedEF, result.EaseFactor)
	}
}
