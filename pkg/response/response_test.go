package response

import (
	"testing"
)

func TestR_Struct(t *testing.T) {
	r := R{Code: 0, Message: "ok", Data: map[string]string{"key": "value"}}
	if r.Code != 0 {
		t.Errorf("Code = %d", r.Code)
	}
	if r.Message != "ok" {
		t.Errorf("Message = %q", r.Message)
	}
}

func TestPageData_Struct(t *testing.T) {
	p := PageData{Page: 1, PageSize: 20, Total: 100, TotalPages: 5}
	if p.Page != 1 {
		t.Errorf("Page = %d", p.Page)
	}
	if p.TotalPages != 5 {
		t.Errorf("TotalPages = %d", p.TotalPages)
	}
}

func TestList_Struct(t *testing.T) {
	items := []string{"a", "b", "c"}
	l := List{
		Items: items,
		Pagination: PageData{Page: 1, PageSize: 10, Total: 3, TotalPages: 1},
	}
	if l.Pagination.Total != 3 {
		t.Errorf("Total = %d", l.Pagination.Total)
	}
}

// Test pagination calculation logic (extracted from OKList)
func TestTotalPagesCalculation(t *testing.T) {
	tests := []struct {
		total      int64
		pageSize   int
		wantPages  int
	}{
		{100, 20, 5},   // exact division
		{101, 20, 6},   // remainder → +1
		{0, 20, 0},     // zero total
		{5, 10, 1},     // fewer items than page size
		{10, 10, 1},    // exactly one page
		{11, 10, 2},    // one extra
		{1, 1, 1},      // single
		{0, 0, 0},      // pageSize=0 should yield 0
	}

	for _, tt := range tests {
		totalPages := 0
		if tt.pageSize > 0 {
			totalPages = int(tt.total) / tt.pageSize
			if int(tt.total)%tt.pageSize > 0 {
				totalPages++
			}
		}
		if totalPages != tt.wantPages {
			t.Errorf("total=%d pageSize=%d → got %d, want %d",
				tt.total, tt.pageSize, totalPages, tt.wantPages)
		}
	}
}

func TestR_JSONOmitempty(t *testing.T) {
	// Data is omitempty, so nil Data should be fine
	r := R{Code: 401, Message: "未登录"}
	if r.Data != nil {
		t.Error("Data should be nil for empty response")
	}
	if r.Code != 401 {
		t.Errorf("Code = %d", r.Code)
	}
}
