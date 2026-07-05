package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

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

// =========================
// MAIN MIDDLEWARE
// =========================
func MiddleManAPI() gin.HandlerFunc {
	return func(c *gin.Context) {

		// =========================
		// 1. LIMIT REQUEST SIZE (ANTI-DoS)
		// =========================
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1MB

		// =========================
		// 2. READ BODY SAFELY
		// =========================
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// restore body for controller
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// =========================
		// 3. BUILD SCAN PAYLOAD
		// =========================
		scanPayload := ScanRequest{
			IP:      c.ClientIP(),
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Headers: c.Request.Header,
			Body:    string(bodyBytes),
		}

		// =========================
		// 4. CALL AI MIDDLEMAN
		// =========================
		result, err := sendJsonToMiddleman(scanPayload)
		if err != nil {
			c.AbortWithStatusJSON(503, gin.H{
				"error":   "security scanner unavailable",
				"blocked": true,
			})
			return
		}

		// =========================
		// 5. VALIDATE RESPONSE (ANTI-BYPASS)
		// =========================
		if result == nil {
			c.AbortWithStatusJSON(503, gin.H{
				"error":   "invalid security response",
				"blocked": true,
			})
			return
		}

		// =========================
		// 6. BLOCK IF MALICIOUS
		// =========================
		if result.Suspicious {
			c.AbortWithStatusJSON(403, gin.H{
				"error": "malicious request detected",
				"risk":  result.Probability,
			})
			return
		}

		// =========================
		// 7. CONTINUE REQUEST
		// =========================
		c.Next()
	}
}

// =========================
// CALL MIDDLEMAN API
// =========================
func sendJsonToMiddleman(scan ScanRequest) (*ScanResponse, error) {

	url := os.Getenv("middelmanware_url") + "/scan"

	jsonData, err := json.Marshal(scan)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 3 * time.Second, // strict security timeout
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// =========================
	// 1. CHECK HTTP STATUS
	// =========================
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("middleman error: status %d", resp.StatusCode)
	}

	// =========================
	// 2. DECODE RESPONSE SAFELY
	// =========================
	var result ScanResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
