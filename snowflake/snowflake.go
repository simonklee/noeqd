package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	sequenceBits = uint64(12)
	workerIdBits = uint64(10)

	SequenceMax = int64(-1) ^ (int64(-1) << sequenceBits)
	WorkerIdMax = int64(-1) ^ (int64(-1) << workerIdBits)

	workerIdShift  = sequenceBits
	timestampShift = sequenceBits + workerIdBits

	// 1 Jan 2012 00:00:00.000 GMT
	epoch = int64(1325289600000)
)

type Snowflake struct {
	lastTimestamp int64
	workerId      int64
	sequence      int64
	mu            sync.Mutex
}

func (sf *Snowflake) id() int64 {
	return (sf.lastTimestamp << timestampShift) |
		(sf.workerId << workerIdShift) |
		(sf.sequence)
}

func New(workerId int64) (*Snowflake, error) {
	if workerId < 0 || workerId > WorkerIdMax {
		return nil, fmt.Errorf("Worker id %v is invalid", workerId)
	}

	return &Snowflake{workerId: workerId}, nil
}

func (sf *Snowflake) Next() (int64, error) {
	sf.mu.Lock()

	ts := timestamp()

	if ts < sf.lastTimestamp {
		return 0, fmt.Errorf("Time is moving backwards, waiting until %d\n", sf.lastTimestamp)
	}

	if ts == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & SequenceMax

		if sf.sequence == 0 {
			ts = nextTimestamp(ts)
		}
	} else {
		sf.sequence = 0
	}

	sf.lastTimestamp = ts
	id := sf.id()
	sf.mu.Unlock()
	return id, nil
}

func timestamp() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - epoch
}

func nextTimestamp(prev int64) int64 {
	ts := timestamp()

	// wait 1 ms
	for ts <= prev {
		time.Sleep(time.Microsecond * 100)
		ts = timestamp()
	}

	return ts
}
