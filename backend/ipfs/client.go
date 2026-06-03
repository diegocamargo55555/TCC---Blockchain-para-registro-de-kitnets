package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const ipfsNodeURL = "http://localhost:5001/api/v0"

// AddFileToIPFS uploads a file byte array to IPFS and returns its CID Hash
func AddFileToIPFS(filename string, fileData []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	part.Write(fileData)
	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", ipfsNodeURL+"/add?cid-version=1", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha ao conectar no IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IPFS retornou status HTTP %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", err
	}

	hash, ok := result["Hash"].(string)
	if !ok {
		return "", fmt.Errorf("resposta invalida do IPFS: %s", string(respBody))
	}

	return hash, nil
}
