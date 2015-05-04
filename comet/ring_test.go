package main

import (
	"testing"
)

func TestRing(t *testing.T) {
	r := NewRing(3)
	p0, err := r.Set()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	p0.SeqId = 10
	r.SetAdv()
	p1, err := r.Set()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	p1.SeqId = 11
	r.SetAdv()
	p2, err := r.Set()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	p2.SeqId = 12
	r.SetAdv()
	p3, err := r.Set()
	if err != ErrRingFull || p3 != nil {
		t.Error(err)
		t.FailNow()
	}
	p0, err = r.Get()
	if err != nil && p0.SeqId != 10 {
		t.Error(err)
		t.FailNow()
	}
	r.GetAdv()
	p1, err = r.Get()
	if err != nil && p1.SeqId != 11 {
		t.Error(err)
		t.FailNow()
	}
	r.GetAdv()
	p2, err = r.Get()
	if err != nil && p2.SeqId != 12 {
		t.Error(err)
		t.FailNow()
	}
	r.GetAdv()
	p3, err = r.Get()
	if err != ErrRingEmpty || p3 != nil {
		t.Error(err)
		t.FailNow()
	}
	p0, err = r.Set()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	p0.SeqId = 10
	r.SetAdv()
	p1, err = r.Set()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	p1.SeqId = 11
	r.SetAdv()
	p2, err = r.Set()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	p2.SeqId = 12
	r.SetAdv()
	p3, err = r.Set()
	if err != ErrRingFull || p3 != nil {
		t.Error(err)
		t.FailNow()
	}
	p0, err = r.Get()
	if err != nil && p0.SeqId != 10 {
		t.Error(err)
		t.FailNow()
	}
	r.GetAdv()
	p1, err = r.Get()
	if err != nil && p1.SeqId != 11 {
		t.Error(err)
		t.FailNow()
	}
	r.GetAdv()
	p2, err = r.Get()
	if err != nil && p2.SeqId != 12 {
		t.Error(err)
		t.FailNow()
	}
	r.GetAdv()
	p3, err = r.Get()
	if err != ErrRingEmpty || p3 != nil {
		t.Error(err)
		t.FailNow()
	}
}
