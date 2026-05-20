package utils

import (
	"sort"
	"time"

	"k8soperation/internal/app/models"
)

func TryEventsV1First(supportsV1 bool) bool {
	return supportsV1
}

func ApplySinceAndSort(list []models.EventItem, sinceSeconds int64) []models.EventItem {
	if sinceSeconds > 0 {
		cutoff := time.Now().Add(-time.Duration(sinceSeconds) * time.Second)
		n := 0
		for _, it := range list {
			if it.EventTime.IsZero() || it.EventTime.After(cutoff) {
				list[n] = it
				n++
			}
		}
		list = list[:n]
	}

	sort.SliceStable(list, func(i, j int) bool {
		ti, tj := list[i].EventTime, list[j].EventTime
		if ti.IsZero() && tj.IsZero() {
			return false
		}
		if ti.IsZero() {
			return false
		}
		if tj.IsZero() {
			return true
		}
		return ti.After(tj)
	})

	return list
}
