package limiter

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAllow(t *testing.T) {
	l, stop := NewGlobalLimiter(2, time.Second)
	defer stop()

	assert.True(t, l.Allow())
	assert.True(t, l.Allow())
	assert.False(t, l.Allow())
}
