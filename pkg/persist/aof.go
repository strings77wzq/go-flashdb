package persist

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

type AOFMode int

const (
	AOFAlways AOFMode = iota
	AOFEverysec
	AOFNo
)

type AOFPersister struct {
	filename   string
	file       *os.File
	writer     *bufio.Writer
	aofChan    chan []byte
	mode       AOFMode
	mu         sync.Mutex
	lastFsync  time.Time
	shutdownCh chan struct{}
}

func NewAOFPersister(filename string, mode AOFMode) (*AOFPersister, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	p := &AOFPersister{
		filename:   filename,
		file:       file,
		writer:     bufio.NewWriter(file),
		aofChan:    make(chan []byte, 1<<20),
		mode:       mode,
		shutdownCh: make(chan struct{}),
	}

	go p.writeLoop()
	return p, nil
}

func (p *AOFPersister) writeLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case cmd := <-p.aofChan:
			p.writeCmd(cmd)
			if p.mode == AOFAlways {
				p.fsync()
			}
		case <-ticker.C:
			if p.mode == AOFEverysec {
				p.fsync()
			}
		case <-p.shutdownCh:
			p.flushAndClose()
			return
		}
	}
}

func (p *AOFPersister) writeCmd(cmd []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writer.Write(cmd)
	p.writer.Flush()
}

func (p *AOFPersister) fsync() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writer.Flush()
	p.file.Sync()
	p.lastFsync = time.Now()
}

func (p *AOFPersister) Append(cmd []byte) {
	p.aofChan <- append([]byte(nil), cmd...)
}

func (p *AOFPersister) flushAndClose() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writer.Flush()
	p.file.Sync()
	p.file.Close()
}

func (p *AOFPersister) Close() {
	close(p.shutdownCh)
}

func (p *AOFPersister) Load() ([][]byte, error) {
	file, err := os.Open(p.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var commands [][]byte
	reader := bufio.NewReader(file)
	buf := make([]byte, 0, 256)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		buf = append(buf, b)
		if b == '\n' && len(buf) >= 2 && buf[len(buf)-2] == '\r' {
			commands = append(commands, append([]byte(nil), buf...))
			buf = buf[:0]
		}
	}

	return commands, nil
}

func (p *AOFPersister) Rewrite(data map[string][]byte) error {
	tmpFile := p.filename + ".tmp"
	file, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)

	for key, value := range data {
		cmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(key), key, len(value), string(value))
		writer.WriteString(cmd)
	}
	writer.Flush()
	file.Sync()
	file.Close()

	os.Rename(tmpFile, p.filename)
	return nil
}
