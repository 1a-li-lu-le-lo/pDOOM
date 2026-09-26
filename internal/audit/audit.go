// Copyright NU Cybernetics. p(DOOM) — research prototype.

// Package audit implements the append-only, hash-chained audit log at
// data/audit/audit.jsonl. Every event carries the hash of its predecessor;
// Verify recomputes the chain and fails on any edit, deletion or reordering.
package audit

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/1a-li-lu-le-lo/pdoom/internal/schema"
)

// Event is one audit record.
type Event struct {
	Seq      int            `json:"seq"`
	TS       string         `json:"ts"`
	Actor    string         `json:"actor"`
	Action   string         `json:"action"`
	Subject  string         `json:"subject"`
	Details  map[string]any `json:"details"`
	PrevHash string         `json:"prev_hash"`
	Hash     string         `json:"hash"`
}

// genesisHash is the prev_hash of the first event.
const genesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// hashOf computes sha256(prev_hash + canonical_json(event without hash)).
func hashOf(ev Event) (string, error) {
	ev.Hash = ""
	body, err := schema.CanonicalJSON(ev)
	if err != nil {
		return "", err
	}
	return schema.SHA256Hex(append([]byte(ev.PrevHash), body...)), nil
}

// Read loads every event of the log. A missing file yields an empty log.
func Read(path string) ([]Event, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var events []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	line := 0
	for sc.Scan() {
		line++
		if len(sc.Bytes()) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			return nil, fmt.Errorf("audit: line %d: %w", line, err)
		}
		events = append(events, ev)
	}
	return events, sc.Err()
}

// Append adds an event to the log, filling seq, prev_hash and hash. ts, actor,
// action and subject are supplied by the caller (the log never reads the clock).
func Append(path string, ev Event) (Event, error) {
	if ev.TS == "" || ev.Actor == "" || ev.Action == "" {
		return ev, fmt.Errorf("audit: ts, actor and action are required")
	}
	events, err := Read(path)
	if err != nil {
		return ev, err
	}
	if err := verify(events); err != nil {
		return ev, fmt.Errorf("audit: refusing to append to a broken chain: %w", err)
	}
	ev.Seq = len(events) + 1
	ev.PrevHash = genesisHash
	if len(events) > 0 {
		ev.PrevHash = events[len(events)-1].Hash
	}
	if ev.Details == nil {
		ev.Details = map[string]any{}
	}
	h, err := hashOf(ev)
	if err != nil {
		return ev, err
	}
	ev.Hash = h
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return ev, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return ev, err
	}
	defer f.Close()
	raw, err := schema.CanonicalJSON(ev)
	if err != nil {
		return ev, err
	}
	if _, err := f.Write(append(raw, '\n')); err != nil {
		return ev, err
	}
	return ev, nil
}

// Verify checks the whole chain on disk.
func Verify(path string) (int, error) {
	events, err := Read(path)
	if err != nil {
		return 0, err
	}
	return len(events), verify(events)
}

func verify(events []Event) error {
	prev := genesisHash
	for i, ev := range events {
		if ev.Seq != i+1 {
			return fmt.Errorf("event %d has seq %d", i+1, ev.Seq)
		}
		if ev.PrevHash != prev {
			return fmt.Errorf("event %d prev_hash does not match predecessor", ev.Seq)
		}
		h, err := hashOf(ev)
		if err != nil {
			return err
		}
		if h != ev.Hash {
			return fmt.Errorf("event %d hash mismatch (record altered)", ev.Seq)
		}
		prev = ev.Hash
	}
	return nil
}
