package world_test

import (
	"math"
	"testing"

	"github.com/herogame/backend/internal/world"
)

func TestUpkeepSlowdown_table(t *testing.T) {
	cases := []struct {
		army int
		want float64
	}{
		{0, 1.000},
		{1, 1.000},
		{10, 1.000},
		{20, 2.000},
		{50, 3.322},
		{100, 4.322},
		{200, 5.322},
		{500, 6.644},
		{1000, 7.644},
		{5000, 9.966},
	}
	for _, tc := range cases {
		got := world.UpkeepSlowdown(tc.army)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("UpkeepSlowdown(%d) = %.3f, want %.3f", tc.army, got, tc.want)
		}
	}
}

func TestTravelSeconds_table(t *testing.T) {
	const baseSpeed = 10
	cases := []struct {
		dist, army, want int
	}{
		{20, 0, 2},
		{20, 50, 7},
		{30, 0, 3},
		{30, 50, 10},
		{30, 200, 16},
		{30, 1000, 23},
	}
	for _, tc := range cases {
		got := world.TravelSeconds(tc.dist, baseSpeed, tc.army)
		if got != tc.want {
			t.Errorf("TravelSeconds(%d, %d, %d) = %d, want %d", tc.dist, baseSpeed, tc.army, got, tc.want)
		}
	}
}
