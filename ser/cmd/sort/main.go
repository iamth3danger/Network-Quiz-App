package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Row struct {
	Name       string
	Streak     int32
	RightCount int32
	WrongCount int32
}

func sortRows(rows []Row) {
	slices.SortFunc(rows, func(a, b Row) int {
		if n := cmp.Compare(a.Streak, b.Streak); n != 0 {
			return n
		}
		aRatio := float64(a.RightCount) / max(float64(a.WrongCount), 1)
		bRatio := float64(b.RightCount) / max(float64(b.WrongCount), 1)
		return cmp.Compare(aRatio, bRatio)
	})
}

func main() {
	rows := []Row{
		{Name: "A", Streak: 2, RightCount: 10, WrongCount: 2}, // ratio 5.0
		{Name: "B", Streak: 0, RightCount: 1, WrongCount: 5},  // ratio 0.2
		{Name: "C", Streak: 1, RightCount: 3, WrongCount: 3},  // ratio 1.0
		{Name: "D", Streak: 0, RightCount: 8, WrongCount: 2},  // ratio 4.0
		{Name: "E", Streak: 2, RightCount: 1, WrongCount: 4},  // ratio 0.25
		{Name: "F", Streak: 1, RightCount: 1, WrongCount: 10}, // ratio 0.1
		{Name: "G", Streak: 0, RightCount: 2, WrongCount: 2},  // ratio 1.0
	}

	fmt.Println("Before:")
	for _, r := range rows {
		ratio := float64(r.RightCount) / max(float64(r.WrongCount), 1)
		fmt.Printf("  %s  streak=%-2d  right=%-2d  wrong=%-2d  ratio=%.2f\n", r.Name, r.Streak, r.RightCount, r.WrongCount, ratio)
	}

	sortRows(rows)

	fmt.Println("\nAfter:")
	for _, r := range rows {
		ratio := float64(r.RightCount) / max(float64(r.WrongCount), 1)
		fmt.Printf("  %s  streak=%-2d  right=%-2d  wrong=%-2d  ratio=%.2f\n", r.Name, r.Streak, r.RightCount, r.WrongCount, ratio)
	}
}
