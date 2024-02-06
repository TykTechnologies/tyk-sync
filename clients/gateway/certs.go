package gateway

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/ongoingio/urljoin"
)

func (c *Client) CreateCertificate(cert []byte) (string, error) {
	fullPath := urljoin.Join(c.url, endpointCerts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("cert", "cert.pem")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, io.NopCloser(bytes.NewReader(cert)))

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fullPath, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Tyk-Authorization", c.secret)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)

	rBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API Returned error: %v", string(rBody))
	}

	dbResp := objects.CertResponse{}
	if err := json.Unmarshal(rBody, &dbResp); err != nil {
		return "", err
	}

	if strings.ToLower(dbResp.Status) != "ok" {
		return "", fmt.Errorf("API request completed, but with error: %v", dbResp.Message)
	}

	return dbResp.Id, nil
}
