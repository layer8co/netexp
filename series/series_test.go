package series_test

import (
	"testing"
	"netexp/series"
	"strconv"
)

func TestRecord(t *testing.T) {
	keep := 10
	ts := series.New[int](keep)

	for i := 0; i <= 20; i++ {
		ts.Record(i)
	}

	got := ts.Samples[len(ts.Samples) - 1]
	want := 20

	if got != want {
		t.Errorf("incorrect last sample; got %d, want %d", got, want)
	}
}

func TestKeep(t *testing.T) {
	keep := 10
	ts := series.New[int](keep)

	for i := 0; i <= 20; i++ {
		ts.Record(i)
	}

	got := len(ts.Samples)
	want := keep

	if got != want {
		t.Errorf("incorrect sample count; got %d, want %d", got, want)
	}
}

func TestLast(t *testing.T) {
	keep := 10
	ts := series.New[int](keep)

	for i := 0; i <= 20; i++ {
		ts.Record(i)
	}

	got := ts.Last(3)
	want := 18

	if got != want {
		t.Errorf("incorrect last; got %d, want %d", got, want)
	}
}

func TestHead(t *testing.T) {
	keep := 10
	ts := series.New[int](keep)

	for i := 0; i <= 20; i++ {
		ts.Record(i)
	}

	got := len(ts.Head(8))
	want := 8

	if got != want {
		t.Errorf("incorrect head count; got %d, want %d", got, want)
	}
}

func TestMap(t *testing.T) {
	keep := 10
	ts := series.New[int](keep)

	for i := 0; i <= 20; i++ {
		ts.Record(i)
	}

	tostring := func(i int) string {
		return strconv.Itoa(i)
	}
	newts := series.Map(ts, tostring)

	got := newts.Samples[len(newts.Samples) - 1]
	want := "20"

	if got != want {
		t.Errorf("incorrect mapped series sample; got %s, want %s", got, want)
	}
}
