package world

import "math"

// SegmentDistance returns minimum distance between two line segments.
func SegmentDistance(ax, ay, bx, by, cx, cy, dx, dy float64) float64 {
	const samples = 10
	min := math.MaxFloat64
	for i := 0; i <= samples; i++ {
		t := float64(i) / float64(samples)
		px := ax + (bx-ax)*t
		py := ay + (by-ay)*t
		for j := 0; j <= samples; j++ {
			u := float64(j) / float64(samples)
			qx := cx + (dx-cx)*u
			qy := cy + (dy-cy)*u
			dx := px - qx
			dy := py - qy
			d := math.Sqrt(dx*dx + dy*dy)
			if d < min {
				min = d
			}
		}
	}
	return min
}
