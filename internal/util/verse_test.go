package util

import (
	"testing"
)

func TestParseVerseReference(t *testing.T) {
	tests := []struct {
		ref         string
		wantBook    string
		wantChapter string
		wantVerse   string
		wantErr     bool
	}{
		{
			ref:         "John 3:16",
			wantBook:    "John",
			wantChapter: "3",
			wantVerse:   "16",
			wantErr:     false,
		},
		{
			ref:         "1 John 1:9",
			wantBook:    "1 John",
			wantChapter: "1",
			wantVerse:   "9",
			wantErr:     false,
		},
		{
			ref:         "Psalm 23",
			wantBook:    "Psalm",
			wantChapter: "23",
			wantVerse:   "",
			wantErr:     false,
		},
		{
			ref:         "Song of Solomon 2:1",
			wantBook:    "Song of Solomon",
			wantChapter: "2",
			wantVerse:   "1",
			wantErr:     false,
		},
		{
			ref:         "InvalidReference",
			wantBook:    "",
			wantChapter: "",
			wantVerse:   "",
			wantErr:     true,
		},
		{
			ref:         "John :16",
			wantBook:    "",
			wantChapter: "",
			wantVerse:   "",
			wantErr:     true,
		},
		{
			ref:         "John 1:12-2:4",
			wantBook:    "John",
			wantChapter: "1",
			wantVerse:   "12-2:4",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			gotBook, gotChapter, gotVerse, err := ParseVerseReference(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVerseReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBook != tt.wantBook {
				t.Errorf("ParseVerseReference() gotBook = %v, want %v", gotBook, tt.wantBook)
			}
			if gotChapter != tt.wantChapter {
				t.Errorf("ParseVerseReference() gotChapter = %v, want %v", gotChapter, tt.wantChapter)
			}
			if gotVerse != tt.wantVerse {
				t.Errorf("ParseVerseReference() gotVerse = %v, want %v", gotVerse, tt.wantVerse)
			}
		})
	}
}

func TestParseVerseRange(t *testing.T) {
	tests := []struct {
		verseStr string
		want     ParsedVerseRange
		wantErr  bool
	}{
		{
			verseStr: "16",
			want: ParsedVerseRange{
				StartVerse:     16,
				EndVerse:       16,
				EndChapter:     0,
				IsCrossChapter: false,
			},
			wantErr: false,
		},
		{
			verseStr: "16-20",
			want: ParsedVerseRange{
				StartVerse:     16,
				EndVerse:       20,
				EndChapter:     0,
				IsCrossChapter: false,
			},
			wantErr: false,
		},
		{
			verseStr: "12-2:4",
			want: ParsedVerseRange{
				StartVerse:     12,
				EndVerse:       4,
				EndChapter:     2,
				IsCrossChapter: true,
			},
			wantErr: false,
		},
		{
			verseStr: "invalid",
			want:     ParsedVerseRange{},
			wantErr:  true,
		},
		{
			verseStr: "12-invalid",
			want:     ParsedVerseRange{},
			wantErr:  true,
		},
		{
			verseStr: "12-2:invalid",
			want:     ParsedVerseRange{},
			wantErr:  true,
		},
		{
			verseStr: "",
			want:     ParsedVerseRange{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.verseStr, func(t *testing.T) {
			got, err := ParseVerseRange(tt.verseStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVerseRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseVerseRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
