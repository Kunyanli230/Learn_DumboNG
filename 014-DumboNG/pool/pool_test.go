package pool

import (
	"learn_DumboNG/014-DumboNG/logger"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "info.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	logger.SetOutput(logger.InfoLevel, file)

	params := DefaultParameters
	params.Rate = 100
	params.BatchSize = 10
	pool := NewPool(params, 4, 0)
	pool.Run()
	time.Sleep(200 * time.Millisecond)
	batch := pool.GetBatch()
	if batch.ID == -1 {
		t.Fatal("expected a generated batch")
	}
}
