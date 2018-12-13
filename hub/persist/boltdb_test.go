package persist

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/baidu/openedge/hub/utils"
	"github.com/stretchr/testify/assert"
)

func TestBoltDBPerf(t *testing.T) {
	os.RemoveAll("test_bolt.db")
	defer os.RemoveAll("test_bolt.db")
	db, err := NewBoltDB("test_bolt.db")
	assert.NoError(t, err)
	total := uint64(10000)
	batchSize := uint64(10)
	msgs := make([][]byte, 0)
	for i := uint64(0); i < batchSize; i++ {
		msgs = append(msgs, []byte("test"))
	}
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		for i := uint64(1); i <= total; i = i + batchSize {
			db.BatchPutV(msgs)
		}
	}()
	i := uint64(1)
	for i < total/2 {
		start := time.Now()
		kvs, _ := db.BatchFetch(utils.U64ToB(i), int(batchSize))
		if len(kvs) == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		fmt.Printf("Fetch %d messages elapsed time %v\n", len(kvs), time.Since(start))
		i = i + uint64(len(kvs))
	}
	wg.Wait()
	fmt.Printf("Put finished\n")
	for i <= total {
		start := time.Now()
		kvs, _ := db.BatchFetch(utils.U64ToB(i), int(batchSize))
		if len(kvs) == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		fmt.Printf("Fetch %d messages elapsed time %v\n", len(kvs), time.Since(start))
		i = i + uint64(len(kvs))
	}
}

func BenchmarkBoltDBPut(b *testing.B) {
	db, _ := NewBoltDB("BenchmarkBoltDB.db")
	defer db.Close()
	for i := 0; i < b.N; i++ {
		k := utils.U64ToB(uint64(i))
		db.Put(k, k)
	}
}

func BenchmarkBoltDBGet(b *testing.B) {
	db, _ := NewBoltDB("BenchmarkBoltDB.db")
	defer db.Close()
	for i := 0; i < b.N; i++ {
		k := utils.U64ToB(uint64(i))
		db.Get(k)
	}
}
