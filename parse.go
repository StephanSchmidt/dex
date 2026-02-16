package main

import (
	"fmt"
	"strconv"
	"strings"
)

// splitDirRange splits "dir/range" into ("dir/", "range"). If there is
// no "/" the dir part is empty and the whole string is the range.
func splitDirRange(expr string) (dir, rangeExpr string) {
	i := strings.LastIndex(expr, "/")
	if i < 0 {
		return "", expr
	}
	return expr[:i+1], expr[i+1:]
}

// parseSliceExpr parses a comma-separated list of 1-based slide selectors
// into 0-based indices. Each item is either a single number or a range
// (using : or -). Negative numbers count from the end.
func parseSliceExpr(expr string, length int) ([]int, error) {
	var result []int
	for _, part := range strings.Split(expr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try colon range first, then dash range.
		start, end, isRange := "", "", false
		if idx := strings.Index(part, ":"); idx >= 0 {
			start, end = part[:idx], part[idx+1:]
			isRange = true
		} else {
			// For dash: only treat as range separator when preceded by a digit.
			for j := 1; j < len(part); j++ {
				if part[j] == '-' && part[j-1] >= '0' && part[j-1] <= '9' {
					start, end = part[:j], part[j+1:]
					isRange = true
					break
				}
			}
		}

		if isRange {
			s, err := resolveIndex(start, length)
			if err != nil {
				return nil, fmt.Errorf("invalid range start %q: %w", start, err)
			}
			e, err := resolveIndex(end, length)
			if err != nil {
				return nil, fmt.Errorf("invalid range end %q: %w", end, err)
			}
			if s > e {
				return nil, fmt.Errorf("invalid range: %d > %d", s+1, e+1)
			}
			for i := s; i <= e; i++ {
				result = append(result, i)
			}
		} else {
			idx, err := resolveIndex(part, length)
			if err != nil {
				return nil, err
			}
			result = append(result, idx)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("empty slide expression")
	}
	return result, nil
}

// resolveIndex converts a 1-based (possibly negative) index string to a
// 0-based index. Negative values count from the end (-1 = last).
func resolveIndex(s string, length int) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	if n == 0 {
		return 0, fmt.Errorf("index 0 is not valid (1-based)")
	}
	var idx int
	if n > 0 {
		idx = n - 1
	} else {
		idx = length + n
	}
	if idx < 0 || idx >= length {
		return 0, fmt.Errorf("index %d out of range (1..%d)", n, length)
	}
	return idx, nil
}
