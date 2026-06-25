package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllowsWithinBurst(t *testing.T) {
	l := New(true, 10, 5)
	ip := "192.168.1.1"

	for i := 0; i < 5; i++ {
		if !l.Allow(ip) {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestLimiterBlocksWhenExhausted(t *testing.T) {
	l := New(true, 1, 1)
	ip := "192.168.1.2"

	if !l.Allow(ip) {
		t.Fatal("first request should be allowed")
	}
	if l.Allow(ip) {
		t.Fatal("second immediate request should be blocked")
	}
}

func TestLimiterRefillsOverTime(t *testing.T) {
	l := New(true, 10, 1)
	ip := "192.168.1.3"

	if !l.Allow(ip) {
		t.Fatal("first request should be allowed")
	}
	time.Sleep(150 * time.Millisecond)
	if !l.Allow(ip) {
		t.Fatal("request should be allowed after refill")
	}
}

func TestDisabledLimiter(t *testing.T) {
	l := New(false, 1, 1)
	for i := 0; i < 100; i++ {
		if !l.Allow("any") {
			t.Fatal("disabled limiter should allow all")
		}
	}
}
