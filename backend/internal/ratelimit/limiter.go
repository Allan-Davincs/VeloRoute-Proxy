package ratelimit

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
	lastAccess time.Time
	mu         sync.Mutex
}

// Limiter implements per-IP token bucket rate limiting.
type Limiter struct {
	mu                sync.RWMutex
	buckets           map[string]*tokenBucket
	requestsPerSecond float64
	burst             float64
	enabled           bool
	stopCh            chan struct{}
}

// New creates a new rate limiter.
func New(enabled bool, requestsPerSecond, burst int) *Limiter {
	l := &Limiter{
		buckets:           make(map[string]*tokenBucket),
		requestsPerSecond: float64(requestsPerSecond),
		burst:             float64(burst),
		enabled:           enabled,
		stopCh:            make(chan struct{}),
	}
	if enabled {
		go l.cleanupLoop()
	}
	return l
}

// Allow checks if a request from clientIP should be allowed.
func (l *Limiter) Allow(clientIP string) bool {
	if !l.enabled {
		return true
	}

	l.mu.RLock()
	bucket, ok := l.buckets[clientIP]
	l.mu.RUnlock()

	if !ok {
		l.mu.Lock()
		bucket, ok = l.buckets[clientIP]
		if !ok {
			bucket = &tokenBucket{
				tokens:     l.burst,
				lastRefill: time.Now(),
				lastAccess: time.Now(),
			}
			l.buckets[clientIP] = bucket
		}
		l.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = math.Min(bucket.tokens+elapsed*l.requestsPerSecond, l.burst)
	bucket.lastRefill = now
	bucket.lastAccess = now

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}
	return false
}

// RetryAfterSeconds returns seconds until a token is available for clientIP.
func (l *Limiter) RetryAfterSeconds(clientIP string) int {
	l.mu.RLock()
	bucket, ok := l.buckets[clientIP]
	l.mu.RUnlock()
	if !ok {
		return 1
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if bucket.tokens >= 1.0 {
		return 0
	}
	needed := 1.0 - bucket.tokens
	secs := int(math.Ceil(needed / l.requestsPerSecond))
	if secs < 1 {
		return 1
	}
	return secs
}

// WriteRateLimitResponse writes a 429 response with Retry-After header.
func WriteRateLimitResponse(w http.ResponseWriter, retryAfter int) {
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.stopCh:
			return
		}
	}
}

func (l *Limiter) cleanup() {
	cutoff := time.Now().Add(-5 * time.Minute)
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, bucket := range l.buckets {
		bucket.mu.Lock()
		stale := bucket.lastAccess.Before(cutoff)
		bucket.mu.Unlock()
		if stale {
			delete(l.buckets, ip)
		}
	}
}

// Stop stops the background cleanup goroutine.
func (l *Limiter) Stop() {
	close(l.stopCh)
}
