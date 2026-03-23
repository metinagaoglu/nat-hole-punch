package peer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/metnagaoglu/holepunch/protocol"
)

const maxChunkSize = 512

// FileTransfer handles chunked file sending/receiving over UDP
type FileTransfer struct {
	mu             sync.Mutex
	incomingChunks map[string]map[int][]byte
	incomingMeta   map[string]*protocol.FileHeader
}

func newFileTransfer() *FileTransfer {
	return &FileTransfer{
		incomingChunks: make(map[string]map[int][]byte),
		incomingMeta:   make(map[string]*protocol.FileHeader),
	}
}

// SendFile reads a file and sends it in chunks to a peer
func (ft *FileTransfer) SendFile(conn *net.UDPConn, peer *net.UDPAddr, localAddr string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileName := filepath.Base(filePath)
	totalChunks := (len(data) + maxChunkSize - 1) / maxChunkSize

	header := protocol.FileHeader{
		Name:        fileName,
		Size:        int64(len(data)),
		TotalChunks: totalChunks,
	}
	headerJSON, _ := json.Marshal(header)
	msg, _ := protocol.EncodePeerMessage(protocol.MsgFile, localAddr, string(headerJSON))
	if _, err := conn.WriteTo(msg, peer); err != nil {
		return fmt.Errorf("failed to send file header: %w", err)
	}

	log.Printf("Sending file '%s' (%d bytes, %d chunks) to %s", fileName, len(data), totalChunks, peer)

	for i := 0; i < totalChunks; i++ {
		start := i * maxChunkSize
		end := start + maxChunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := protocol.FileChunk{
			Name:  fileName,
			Index: i,
			Total: totalChunks,
			Data:  base64.StdEncoding.EncodeToString(data[start:end]),
		}

		chunkJSON, _ := json.Marshal(chunk)
		msg, _ := protocol.EncodePeerMessage(protocol.MsgFileChunk, localAddr, string(chunkJSON))
		if _, err := conn.WriteTo(msg, peer); err != nil {
			return fmt.Errorf("failed to send chunk %d: %w", i, err)
		}
	}

	log.Printf("File '%s' sent successfully", fileName)
	return nil
}

func (ft *FileTransfer) handleFileHeader(header *protocol.FileHeader) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	ft.incomingMeta[header.Name] = header
	ft.incomingChunks[header.Name] = make(map[int][]byte)
	log.Printf("Receiving file '%s' (%d bytes, %d chunks)", header.Name, header.Size, header.TotalChunks)
}

// handleFileChunk returns (complete, error)
func (ft *FileTransfer) handleFileChunk(chunk *protocol.FileChunk) (bool, error) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	decoded, err := base64.StdEncoding.DecodeString(chunk.Data)
	if err != nil {
		return false, fmt.Errorf("failed to decode chunk data: %w", err)
	}

	if ft.incomingChunks[chunk.Name] == nil {
		ft.incomingChunks[chunk.Name] = make(map[int][]byte)
	}
	ft.incomingChunks[chunk.Name][chunk.Index] = decoded

	received := len(ft.incomingChunks[chunk.Name])
	log.Printf("Received chunk %d/%d for '%s'", chunk.Index+1, chunk.Total, chunk.Name)

	if received == chunk.Total {
		return true, ft.assembleFile(chunk.Name, chunk.Total)
	}
	return false, nil
}

func (ft *FileTransfer) assembleFile(name string, totalChunks int) error {
	chunks := ft.incomingChunks[name]

	var data []byte
	for i := 0; i < totalChunks; i++ {
		chunk, ok := chunks[i]
		if !ok {
			return fmt.Errorf("missing chunk %d for file '%s'", i, name)
		}
		data = append(data, chunk...)
	}

	outPath := filepath.Join(".", "received_"+name)
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("File '%s' saved as '%s' (%d bytes)", name, outPath, len(data))

	delete(ft.incomingChunks, name)
	delete(ft.incomingMeta, name)

	return nil
}
