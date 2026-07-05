package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type RequestLog struct {
	IP        string              `json:"ip"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
	UserID    any                 `json:"user_id"`
	UserEmail string              `json:"user_email"`
	Time      time.Time           `json:"time"`
}

type ScanRequest struct {
	IP      string              `json:"ip"`
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type ScanResponse struct {
	Probability float64 `json:"probability"`
	Suspicious  bool    `json:"suspicious"`
}

func MiddleManAPI() gin.HandlerFunc {
	return func(c *gin.Context) {

		// =========================
		// 1. READ BODY SAFELY
		// =========================
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// =========================
		// 2. BUILD SCAN PAYLOAD
		// =========================
		scanPayload := ScanRequest{
			IP:      c.ClientIP(),
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Headers: c.Request.Header,
			Body:    string(bodyBytes),
		}

		// =========================
		// 3. SEND TO MIDDLEMAN API
		// =========================
		result, err := sendJsonToMiddleman(scanPayload)

		// =========================
		// 4. HANDLE TIMEOUT / ERROR (FAIL-CLOSED SECURITY)
		// =========================
		if err != nil {
			c.AbortWithStatusJSON(503, gin.H{
				"error":   "security scanner failed",
				"blocked": true,
			})
			return
		}

		// =========================
		// 5. BLOCK IF DANGEROUS
		// =========================
		if result.Suspicious {
			c.AbortWithStatusJSON(403, gin.H{
				"error": "malicious request detected",
				"risk":  result.Probability,
			})
			return
		}

		// =========================
		// 6. CONTINUE REQUEST
		// =========================
		c.Next()
	}
}

// =========================
// HTTP CALL TO FASTAPI
// =========================
func sendJsonToMiddleman(scan ScanRequest) (*ScanResponse, error) {

	url := os.Getenv("middlemanware_url") + "/scan"

	jsonData, err := json.Marshal(scan)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second, // 🔥 critical security timeout
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err // timeout or network failure
	}
	defer resp.Body.Close()

	var result ScanResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return &result, nil
}
