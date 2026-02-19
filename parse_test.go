package main

import "testing"

func TestParseSliceExpr(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		length  int
		want    []int
		wantErr bool
	}{
		{
			name:   "single",
			expr:   "1",
			length: 15,
			want:   []int{0},
		},
		{
			name:   "comma list",
			expr:   "1,2",
			length: 15,
			want:   []int{0, 1},
		},
		{
			name:   "comma list with spaces",
			expr:   "1, 3",
			length: 15,
			want:   []int{0, 2},
		},
		{
			name:   "dash range",
			expr:   "1-5",
			length: 15,
			want:   []int{0, 1, 2, 3, 4},
		},
		{
			name:   "colon range",
			expr:   "1:5",
			length: 15,
			want:   []int{0, 1, 2, 3, 4},
		},
		{
			name:   "negative end",
			expr:   "1:-1",
			length: 15,
			want:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
		},
		{
			name:   "negative start and end",
			expr:   "-3:-1",
			length: 15,
			want:   []int{12, 13, 14},
		},
		{
			name:   "negative single",
			expr:   "-1",
			length: 15,
			want:   []int{14},
		},
		{
			name:   "last element",
			expr:   "15",
			length: 15,
			want:   []int{14},
		},
		{
			name:    "out of range positive",
			expr:    "16",
			length:  15,
			wantErr: true,
		},
		{
			name:    "out of range negative",
			expr:    "-16",
			length:  15,
			wantErr: true,
		},
		{
			name:    "zero index",
			expr:    "0",
			length:  15,
			wantErr: true,
		},
		{
			name:    "invalid input",
			expr:    "abc",
			length:  15,
			wantErr: true,
		},
		{
			name:    "empty expression",
			expr:    "",
			length:  15,
			wantErr: true,
		},
		{
			name:    "reversed range",
			expr:    "5:1",
			length:  15,
			wantErr: true,
		},
		{
			name:   "colon range with negatives",
			expr:   "-3:-1",
			length: 5,
			want:   []int{2, 3, 4},
		},
		{
			name:   "dash range with negative end",
			expr:   "1--1",
			length: 5,
			want:   []int{0, 1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSliceExpr(tt.expr, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSliceExpr(%q, %d) error = %v, wantErr %v", tt.expr, tt.length, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("parseSliceExpr(%q, %d) = %v, want %v", tt.expr, tt.length, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseSliceExpr(%q, %d) = %v, want %v", tt.expr, tt.length, got, tt.want)
					return
				}
			}
		})
	}
}

func TestResolveInsertIndex(t *testing.T) {
	// 4 slides → valid 1-based positions: 1,2,3,4,5 (5 = append)
	tests := []struct {
		name    string
		expr    string
		nSlides int
		want    int
		wantErr bool
	}{
		{name: "first", expr: "1", nSlides: 4, want: 0},
		{name: "append", expr: "5", nSlides: 4, want: 4},
		{name: "middle", expr: "3", nSlides: 4, want: 2},
		{name: "neg -1 appends", expr: "-1", nSlides: 4, want: 4},
		{name: "neg -2 before last", expr: "-2", nSlides: 4, want: 3},
		{name: "neg -5 first", expr: "-5", nSlides: 4, want: 0},
		{name: "error: zero", expr: "0", nSlides: 4, wantErr: true},
		{name: "error: too high", expr: "6", nSlides: 4, wantErr: true},
		{name: "error: too negative", expr: "-6", nSlides: 4, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveInsertIndex(tt.expr, tt.nSlides)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveInsertIndex(%q, %d) error = %v, wantErr %v", tt.expr, tt.nSlides, err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("resolveInsertIndex(%q, %d) = %d, want %d", tt.expr, tt.nSlides, got, tt.want)
			}
		})
	}
}
