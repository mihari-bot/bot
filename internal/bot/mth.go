package bot

import (
	"math/rand/v2"
	"time"
)

func (b *Bot) mthRandN(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	rangeNs := max.Nanoseconds() - min.Nanoseconds()
	randomNs := rand.Int64N(rangeNs)
	return min + time.Duration(randomNs)
}

func (b *Bot) mthClamp(duration, min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}

	if duration < min {
		return min
	}

	if duration >= max {
		if max == 0 {
			return 0
		}
		return max - 1
	}

	return duration
}

func (b *Bot) mthCalcWaitTime(text string, min, max time.Duration) time.Duration {
	var total time.Duration
	for range text {
		total += b.mthRandN(min, max)
	}
	return total
}
