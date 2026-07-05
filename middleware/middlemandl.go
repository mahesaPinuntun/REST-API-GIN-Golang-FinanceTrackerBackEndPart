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

type FlaggedField struct {
	Path        string  `json:"path"`
	Value       string  `json:"value"`
	Probability float64 `json:"probability"`
	Suspicious  bool    `json:"suspicious"`
}

type ScanResponse struct {
	OverallVerdict   string         `json:"overall_verdict"`
	FieldsScanned    int            `json:"fields_scanned"`
	FieldsFlagged    int            `json:"fields_flagged"`
	HighestRiskScore float64        `json:"highest_risk_score"`
	FlaggedFields    []FlaggedField `json:"flagged_fields"`
}

func MiddleManAPI() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Limit request body (1 MB)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)

		// Read body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "unable to read request body",
			})
			return
		}

		// Restore body for controller
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Nothing to scan on empty bodies (e.g. GET requests) -- skip straight through
		if len(bodyBytes) == 0 {
			c.Next()
			return
		}

		// Send the raw body bytes directly -- no wrapper struct, no
		// re-encoding. The scanner's /scan endpoint expects the actual
		// JSON object as-is so it can walk each field individually.
		result, err := sendToScanner(bodyBytes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "security scanner unavailable",
				"blocked": true,
			})
			return
		}

		if result == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "invalid scanner response",
				"blocked": true,
			})
			return
		}

		// Block request
		if result.OverallVerdict == "SUSPICIOUS" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":          "malicious request detected",
				"risk_score":     result.HighestRiskScore,
				"flagged_fields": result.FlaggedFields,
			})
			return
		}

		c.Next()
	}
}

func sendToScanner(body []byte) (*ScanResponse, error) {
	url := os.Getenv("MIDDLEMANWARE_URL")
	if url == "" {
		return nil, fmt.Errorf("MIDDLEMANWARE_URL is not configured")
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url+"/scan", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scanner returned %d", resp.StatusCode)
	}

	// Read raw response for debugging
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("========== SECURITY SCANNER ==========")
	fmt.Println(string(raw))
	fmt.Println("======================================")

	var result ScanResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
