package inmem

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/cardil/qe-clusterlogging/pkg/clusterlogging"
	"github.com/cardil/qe-clusterlogging/pkg/kubernetes"
	"github.com/cardil/qe-clusterlogging/pkg/storage"
)

type Storage struct {
	mu   sync.RWMutex
	data map[string]*store
}

type store struct {
	info     kubernetes.ContainerInfo
	messages []message
}

func (s store) reader() *messageReader {
	return &messageReader{
		msgNo:    0,
		msgPrt:   0,
		messages: s.messages,
	}
}

type message struct {
	data      []byte
	timestamp time.Time
}

func NewStore() *Storage {
	return &Storage{
		data: make(map[string]*store),
	}
}

func (s *Storage) Store(msg *clusterlogging.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := msg.FullName()
	st, ok := s.data[key]
	if !ok {
		st = &store{
			info:     msg.ContainerInfo,
			messages: make([]message, 0, 1),
		}
		s.data[key] = st
	}
	st.messages = append(st.messages, message{
		data:      []byte(msg.Message),
		timestamp: msg.Timestamp,
	})
	return nil
}

func (s *Storage) Stats() storage.Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	cts := make([]storage.ContainerStat, 0, len(s.data))
	for _, st := range s.data {
		l := len(st.messages)
		cts = append(cts, storage.ContainerStat{
			ContainerInfo: st.info,
			MessageCount:  l,
			LastMessage:   st.messages[l-1].timestamp,
		})
	}
	return cts
}

func (s *Storage) Download() storage.Artifacts {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := map[string]storage.FileReader{}
	for _, st := range s.data {
		key := st.info.FullName()
		data[key+".log"] = func() io.ReadCloser {
			return io.NopCloser(st.reader())
		}
		data[key+".json"] = func() io.ReadCloser {
			by, _ := json.MarshalIndent(st.info, "", "  ")
			return io.NopCloser(bytes.NewReader(by))
		}
	}
	return data
}

type messageReader struct {
	msgNo    int
	msgPrt   int
	messages []message
}

func (mr *messageReader) Read(buf []byte) (int, error) {
	bufLen := len(buf)
	if bufLen == 0 {
		return 0, io.ErrShortBuffer
	}
	j := 0
	for j < bufLen {
		// Check if there are more messages to read, if not, return EOF
		if mr.msgNo >= len(mr.messages) {
			return j, io.EOF
		}
		msg := mr.messages[mr.msgNo]
		dataLen := len(msg.data) - mr.msgPrt
		if dataLen > 0 {
			bufRemain := bufLen - j
			copyable := min(bufRemain, dataLen)
			copy(buf[j:], msg.data[mr.msgPrt:mr.msgPrt+copyable])
			mr.msgPrt += copyable
			j += copyable
		} else {
			mr.msgPrt = 0
			mr.msgNo++
			buf[j] = '\n'
			j++
		}
	}
	return j, nil
}
