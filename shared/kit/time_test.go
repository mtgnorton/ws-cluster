package kit

import "testing"

func TestTimeGetDayByTs(t *testing.T) {
	var ts int64 = 1609434000
	if TimeGetDateByTs(ts) != "2021-01-01" {
		t.Error("TimeGetDateByTs error")
	}
}

func TestTimeGetBeginEndTs(t *testing.T) {
	tests := []struct {
		give []string
		want [2]int64
	}{
		{
			give: []string{"2021-01-01 03:00:00", "2021-02-01 03:00:00"},
			want: [2]int64{1609430400, 1612195199},
		},
		{
			give: []string{"2021-01-1", "2021-2-01"},
			want: [2]int64{1609430400, 1612195199},
		},
		{
			give: []string{"2021-01-01 03:00:00"},
			want: [2]int64{1609430400, 1609516799},
		},
	}
	for _, tt := range tests {
		beginTs, endTs, err := TimeGetBeginEndTs(tt.give...)
		if err != nil {
			t.Error(err)
			continue
		}
		if beginTs != tt.want[0] {
			t.Errorf("TimeGetBeginEndTs beginTs error, want: %d, got: %d", tt.want[0], beginTs)
		}
		if endTs != tt.want[1] {
			t.Errorf("TimeGetBeginEndTs endTs error, want: %d, got: %d", tt.want[1], endTs)
		}

	}
}

func TestTimeDurationDates(t *testing.T) {
	var r = []string{"2021-01-05", "2021-01-04", "2021-01-03", "2021-01-02", "2021-01-01"}

	if dates, err := TimeDurationDates("2021-01-01 01:00:00", "2021-01-05 01:00:00"); err != nil {
		t.Error(err)

	} else {
		if len(dates) != len(r) {
			t.Errorf("TimeDurationDates length error, want: %d, got: %d", len(r), len(dates))
			return
		}
		for i, date := range dates {
			if date != r[i] {
				t.Errorf("TimeDurationDates error, want: %s, got: %s", r[i], date)
			}
		}
	}

}
