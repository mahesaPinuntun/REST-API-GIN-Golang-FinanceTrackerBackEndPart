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

		scanReq := ScanRequest{
			IP:      c.ClientIP(),
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Headers: c.Request.Header,
			Body:    string(bodyBytes),
		}

		result, err := sendToScanner(scanReq)
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

func sendToScanner(scan ScanRequest) (*ScanResponse, error) {

	url := os.Getenv("middelmanware_url")
	if url == "" {
		return nil, fmt.Errorf("MIDDLEMAN_URL is not configured")
	}

	payload, err := json.Marshal(scan)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url+"/scan", bytes.NewBuffer(payload))
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
