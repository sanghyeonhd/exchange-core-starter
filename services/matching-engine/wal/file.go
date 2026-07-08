package wal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// FileLog is a JSON-lines command log. Every append is flushed and fsynced
// before it is acknowledged, so an acknowledged command survives a crash.
type FileLog struct {
	mu   sync.Mutex
	file *os.File
	seq  int64
}

// OpenFileLog opens (or creates) a command log and positions the sequence
// after the last durable record.
func OpenFileLog(path string) (*FileLog, error) {
	existing, err := LoadFileLog(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	log := &FileLog{file: file}
	if len(existing) > 0 {
		log.seq = existing[len(existing)-1].Sequence
	}
	return log, nil
}

func (f *FileLog) Append(cmd Command) (Command, error) {
	if err := validateCommand(cmd); err != nil {
		return Command{}, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	cmd.Sequence = f.seq + 1
	cmd.SchemaVersion = SchemaVersion
	encoded, err := json.Marshal(cmd)
	if err != nil {
		return Command{}, err
	}
	if _, err := f.file.Write(append(encoded, '\n')); err != nil {
		return Command{}, err
	}
	if err := f.file.Sync(); err != nil {
		return Command{}, err
	}
	f.seq = cmd.Sequence
	return cmd, nil
}

func (f *FileLog) LastSequence() int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.seq
}

func (f *FileLog) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}

// LoadFileLog reads all commands from a JSON-lines log and verifies the
// sequence is contiguous and ascending.
func LoadFileLog(path string) ([]Command, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []Command
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var cmd Command
		if err := json.Unmarshal(scanner.Bytes(), &cmd); err != nil {
			return nil, fmt.Errorf("wal line %d: %w", line, err)
		}
		if cmd.SchemaVersion != SchemaVersion {
			return nil, fmt.Errorf("wal line %d: unsupported schema %d", line, cmd.SchemaVersion)
		}
		if err := validateCommand(cmd); err != nil {
			return nil, fmt.Errorf("wal line %d: %w", line, err)
		}
		if want := int64(len(commands) + 1); cmd.Sequence != want {
			return nil, fmt.Errorf("wal line %d: sequence %d, want %d", line, cmd.Sequence, want)
		}
		commands = append(commands, cmd)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return commands, nil
}
