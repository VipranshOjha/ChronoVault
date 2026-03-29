package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptAndStore handles the encryption and shredding logic
func EncryptAndStore(originalData []byte, filename string) (string, string, string, []byte) {
	fmt.Println("--- PHASE 1: ENCRYPT & SHRED ---")

	// 1. Hash Original Data (Identity)
	originalHash := HashData(originalData)
	fmt.Printf("[Enc] Original Hash: %s\n", originalHash[:10])

	// 2. Generate Encryption Key
	key := make([]byte, 32) // AES-256
	io.ReadFull(rand.Reader, key)

	// 3. Encrypt Data
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	encryptedData := gcm.Seal(nonce, nonce, originalData, nil)

	// 4. Shred and Upload Chunks to IPFS
	var chunkHashes []string
	for i := 0; i < len(encryptedData); i += ChunkSize {
		end := i + ChunkSize
		if end > len(encryptedData) {
			end = len(encryptedData)
		}
		chunk := encryptedData[i:end]
		cid, err := UploadChunkToIPFS(chunk, filename)
		if err != nil {
			panic(err)
		}
		chunkHashes = append(chunkHashes, cid)
	}
	fmt.Printf("[Enc] Shredded file into %d chunks\n", len(chunkHashes))

	// 5. Build Merkle Tree
	rootNode := BuildMerkleTree(chunkHashes)
	fmt.Printf("[Enc] Merkle Root Hash: %s...\n", rootNode.Hash[:10])

	// Build Manifest (Order of chunks)
	manifestContent := "# Filename: " + filename + "\n"
	for _, h := range chunkHashes {
		manifestContent += h + "\n"
	}

	return originalHash, rootNode.Hash, manifestContent, key
}
