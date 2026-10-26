package pagination

import "testing"

func TestNew_ClampsValues(t *testing.T) {
	tests := []struct {
		name          string
		page, perPage int
		want          Params
	}{
		{"defaults", 0, 0, Params{Page: 1, PerPage: DefaultPerPage}},
		{"negative", -3, -1, Params{Page: 1, PerPage: DefaultPerPage}},
		{"max per page", 2, 1000, Params{Page: 2, PerPage: MaxPerPage}},
		{"valid", 3, 10, Params{Page: 3, PerPage: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.page, tt.perPage); got != tt.want {
				t.Fatalf("New(%d, %d) = %+v, want %+v", tt.page, tt.perPage, got, tt.want)
			}
		})
	}
}

func TestOffsetAndMeta(t *testing.T) {
	p := New(3, 10)
	if p.Offset() != 20 || p.Limit() != 10 {
		t.Fatalf("offset/limit = %d/%d, want 20/10", p.Offset(), p.Limit())
	}
	m := NewMeta(p, 21)
	if m.TotalPages != 3 {
		t.Fatalf("total pages = %d, want 3", m.TotalPages)
	}
	if NewMeta(p, 0).TotalPages != 0 {
		t.Fatal("total pages for empty result should be 0")
	}
}
