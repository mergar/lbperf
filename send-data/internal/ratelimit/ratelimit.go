package ratelimit

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Limiter struct {
	spoolDir string
}

func New(spoolDir string) *Limiter {
	return &Limiter{spoolDir: spoolDir}
}

func (l *Limiter) Allow(subsystem, clientIP string, maxRecords, intervalSec int) error {
	if maxRecords <= 0 || intervalSec <= 0 {
		return nil
	}

	lockDir := filepath.Join(l.spoolDir, subsystem)
	if err := os.MkdirAll(lockDir, 0700); err != nil {
		return fmt.Errorf("rate limit mkdir: %w", err)
	}

	lockFile := filepath.Join(lockDir, clientIP)
	now := time.Now().Unix()

	count, err := l.pruneAndCount(lockFile, now, int64(intervalSec))
	if err != nil {
		return err
	}
	if count >= maxRecords {
		return ErrRateLimited
	}

	f, err := os.OpenFile(lockFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("rate limit open: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%d\n", now); err != nil {
		return fmt.Errorf("rate limit write: %w", err)
	}
	return nil
}

func (l *Limiter) pruneAndCount(lockFile string, now, intervalSec int64) (int, error) {
	data, err := os.ReadFile(lockFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("rate limit read: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	active := make([]string, 0, len(lines))
	count := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ts, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			continue
		}
		if now-ts < intervalSec {
			active = append(active, line)
			count++
		}
	}

	out := strings.Join(active, "\n")
	if len(active) > 0 {
		out += "\n"
	}
	if err := os.WriteFile(lockFile, []byte(out), 0600); err != nil {
		return 0, fmt.Errorf("rate limit write swap: %w", err)
	}

	return count, nil
}
