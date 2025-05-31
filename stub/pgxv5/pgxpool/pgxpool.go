package pgxpool

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

type Config struct {
	ConnString      string
	MaxConns        int32
	AcquireTimeout  time.Duration
	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

func ParseConfig(dsn string) (*Config, error) {
	return &Config{ConnString: dsn}, nil
}

type Pool struct {
	mu   sync.Mutex
	next int64
	rows map[int64]rowData
}

type rowData struct {
	chatID    int64
	data      string
	createdAt time.Time
}

func NewWithConfig(ctx context.Context, cfg *Config) (*Pool, error) {
	return &Pool{rows: make(map[int64]rowData)}, nil
}

func (p *Pool) QueryRow(ctx context.Context, query string, args ...interface{}) Row {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch {
	case strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "INSERT"):
		chatID := args[0].(int64)
		data := args[1].(string)
		p.next++
		id := p.next
		p.rows[id] = rowData{chatID: chatID, data: data, createdAt: time.Now()}
		return Row{vals: []interface{}{id}}
	case strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT"):
		id := args[0].(int64)
		r, ok := p.rows[id]
		if !ok {
			return Row{err: errors.New("not found")}
		}
		return Row{vals: []interface{}{id, r.chatID, r.data, r.createdAt}}
	default:
		return Row{}
	}
}

func (p *Pool) Exec(ctx context.Context, query string, args ...interface{}) (CommandTag, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "DELETE") {
		id := args[0].(int64)
		delete(p.rows, id)
	}
	return CommandTag{}, nil
}

func (p *Pool) Close() {}

type Row struct {
	vals []interface{}
	err  error
}

func (r Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *int64:
			*d = r.vals[i].(int64)
		case *string:
			*d = r.vals[i].(string)
		case *time.Time:
			*d = r.vals[i].(time.Time)
		}
	}
	return nil
}

type CommandTag struct{}
