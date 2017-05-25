package zapcore

import (
	"net"
	"sync"
)

var _ WriteSyncer = &connSyncer{}

type connSyncer struct {
	conn net.Conn
	dial func() (net.Conn, error)
	m    sync.Mutex
}

type ConnSyncerConfig struct {
	Dial func() (net.Conn, error)
}

func NewConnSyncer(cfg ConnSyncerConfig) (*connSyncer, error) {
	conn, err := cfg.Dial()
	if err != nil {
		return nil, err
	}
	return &connSyncer{
		conn: conn,
		dial: cfg.Dial,
	}, nil
}

func (cs *connSyncer) Write(p []byte) (n int, err error) {
	conn := cs.conn
	if conn != nil {
		if n, err := conn.Write(p); err == nil {
			// optimal path; conn already exists and no error writing; no mutex; net.Conn allows concurrent access
			// each write should be a full message (at least with unix datagram which I'm building for... but may not be true)
			return n, nil
		}
		cs.m.Lock()
		conn.Close()
		cs.conn = nil
		cs.m.Unlock()
	}
	// assume we need to build the conn, but we'll do this without the mutex and discard our conn if a concurrent conn is
	// created.
	conn, err = cs.dial()
	if err != nil {
		return 0, err
	}
	cs.m.Lock()
	if cs.conn != nil {
		conn.Close()
		conn = cs.conn
	} else {
		cs.conn = conn
	}
	cs.m.Unlock()
	return conn.Write(p)
}

func (cs *connSyncer) Sync() error {
	return nil
}
