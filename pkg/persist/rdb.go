package persist

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	RDBMagicString = "REDIS"
	RDBVersion     = "0009"
)

type RDBEncoder struct {
	buffer []byte
}

func NewRDBEncoder() *RDBEncoder {
	return &RDBEncoder{
		buffer: make([]byte, 0, 4096),
	}
}

func (e *RDBEncoder) EncodeHeader() []byte {
	header := append([]byte(RDBMagicString), []byte(RDBVersion)...)
	return header
}

func (e *RDBEncoder) EncodeString(key string, value []byte, expireAt int64) []byte {
	buf := make([]byte, 0, 256+len(key)+len(value))

	if expireAt > 0 {
		buf = append(buf, 0xFC)
		buf = binary.LittleEndian.AppendUint64(buf, uint64(expireAt))
	}

	buf = append(buf, 0x00)
	buf = e.encodeLength(buf, len(key))
	buf = append(buf, key...)
	buf = e.encodeLength(buf, len(value))
	buf = append(buf, value...)

	return buf
}

func (e *RDBEncoder) encodeLength(buf []byte, length int) []byte {
	if length < 1<<6 {
		buf = append(buf, byte(length))
	} else if length < 1<<14 {
		buf = append(buf, byte(length>>8|0x40), byte(length))
	} else {
		buf = append(buf, 0x80)
		buf = binary.LittleEndian.AppendUint32(buf, uint32(length))
	}
	return buf
}

func (e *RDBEncoder) EncodeFooter() []byte {
	return []byte{0xFF}
}

type RDBSaver struct {
	filename string
	encoder  *RDBEncoder
}

func NewRDBSaver(filename string) *RDBSaver {
	return &RDBSaver{
		filename: filename,
		encoder:  NewRDBEncoder(),
	}
}

func (s *RDBSaver) Save(data map[string][]byte, expireTimes map[string]int64) error {
	tmpFile := s.filename + ".tmp"
	file, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	_, _ = file.Write(s.encoder.EncodeHeader())

	for key, value := range data {
		expireAt := int64(0)
		if et, ok := expireTimes[key]; ok {
			expireAt = et
		}
		encoded := s.encoder.EncodeString(key, value, expireAt)
		_, _ = file.Write(encoded)
	}

	_, _ = file.Write(s.encoder.EncodeFooter())
	_ = file.Sync()
	_ = file.Close()

	return os.Rename(tmpFile, s.filename)
}

type RDBLoader struct {
	filename string
}

func NewRDBLoader(filename string) *RDBLoader {
	return &RDBLoader{
		filename: filename,
	}
}

type KVPair struct {
	Key      string
	Value    []byte
	ExpireAt int64
}

func (l *RDBLoader) Load() ([]KVPair, error) {
	file, err := os.Open(l.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	buf := make([]byte, 9)
	_, err = io.ReadFull(file, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read RDB header: %v", err)
	}

	if string(buf[:5]) != RDBMagicString {
		return nil, fmt.Errorf("invalid RDB magic string")
	}

	var pairs []KVPair
	var expireAt int64

	for {
		opcode := make([]byte, 1)
		_, err := file.Read(opcode)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch opcode[0] {
		case 0xFC:
			expireBuf := make([]byte, 8)
			_, err := io.ReadFull(file, expireBuf)
			if err != nil {
				return nil, err
			}
			expireAt = int64(binary.LittleEndian.Uint64(expireBuf))
			continue
		case 0xFF:
			return pairs, nil
		case 0x00:
			key, err := l.readString(file)
			if err != nil {
				return nil, err
			}
			valueStr, err := l.readString(file)
			if err != nil {
				return nil, err
			}
			pairs = append(pairs, KVPair{
				Key:      key,
				Value:    []byte(valueStr),
				ExpireAt: expireAt,
			})
			expireAt = 0
		default:
			return nil, fmt.Errorf("unknown opcode: %x", opcode[0])
		}
	}

	return pairs, nil
}

func (l *RDBLoader) readString(r io.Reader) (string, error) {
	lenBuf := make([]byte, 1)
	_, err := r.Read(lenBuf)
	if err != nil {
		return "", err
	}

	var length int
	switch {
	case lenBuf[0] < 1<<6:
		length = int(lenBuf[0])
	case lenBuf[0] < 1<<7:
		lenBuf2 := make([]byte, 1)
		_, err := r.Read(lenBuf2)
		if err != nil {
			return "", err
		}
		length = int(lenBuf[0]&0x3F)<<8 | int(lenBuf2[0])
	default:
		lenBuf4 := make([]byte, 4)
		_, err := io.ReadFull(r, lenBuf4)
		if err != nil {
			return "", err
		}
		length = int(binary.LittleEndian.Uint32(lenBuf4))
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func (l *RDBLoader) Exists() bool {
	_, err := os.Stat(l.filename)
	return err == nil
}

type PersistManager struct {
	aof          *AOFPersister
	rdbSaver     *RDBSaver
	rdbLoader    *RDBLoader
	aofEnabled   bool
	saveInterval time.Duration
	shutdownCh   chan struct{}
}

func NewPersistManager(aofFile, rdbFile string, aofEnabled bool, saveInterval time.Duration) (*PersistManager, error) {
	var aof *AOFPersister
	var err error

	if aofEnabled {
		aof, err = NewAOFPersister(aofFile, AOFEverysec)
		if err != nil {
			return nil, err
		}
	}

	return &PersistManager{
		aof:          aof,
		rdbSaver:     NewRDBSaver(rdbFile),
		rdbLoader:    NewRDBLoader(rdbFile),
		aofEnabled:   aofEnabled,
		saveInterval: saveInterval,
		shutdownCh:   make(chan struct{}),
	}, nil
}

func (m *PersistManager) AppendAOF(cmd []byte) {
	if m.aofEnabled && m.aof != nil {
		m.aof.Append(cmd)
	}
}

func (m *PersistManager) SaveRDB(data map[string][]byte, expireTimes map[string]int64) error {
	return m.rdbSaver.Save(data, expireTimes)
}

func (m *PersistManager) LoadRDB() ([]KVPair, error) {
	return m.rdbLoader.Load()
}

func (m *PersistManager) LoadAOF() ([][]byte, error) {
	if !m.aofEnabled || m.aof == nil {
		return nil, nil
	}
	return m.aof.Load()
}

func (m *PersistManager) Close() {
	if m.aof != nil {
		m.aof.Close()
	}
	close(m.shutdownCh)
}

func (m *PersistManager) StartAutoSave(dataChan <-chan map[string][]byte, expireChan <-chan map[string]int64) {
	go func() {
		ticker := time.NewTicker(m.saveInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case data := <-dataChan:
					select {
					case expireTimes := <-expireChan:
						m.SaveRDB(data, expireTimes)
					default:
						m.SaveRDB(data, make(map[string]int64))
					}
				default:
				}
			case <-m.shutdownCh:
				return
			}
		}
	}()
}
