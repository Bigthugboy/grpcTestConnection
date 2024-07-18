package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	pb "grpcTestConnection/server/payment/server/payment"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var bearerToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ik1HTHFqOThWTkxvWGFGZnBKQ0JwZ0I0SmFLcyJ9.eyJhdWQiOiI4NzZjY2EyNy02ZTIzLTQyMTgtYmUxOC04N2ExZDJkNGYxOTUiLCJpc3MiOiJodHRwczovL2xvZ2luLm1pY3Jvc29mdG9ubGluZS5jb20vMjc5YzdiMWItYmEwNi00MjdiLWE2ODEtYzhhNTQ5MmQyOTNkL3YyLjAiLCJpYXQiOjE3MjEyNjQ2MzIsIm5iZiI6MTcyMTI2NDYzMiwiZXhwIjoxNzIxMjY4NTMyLCJhaW8iOiJBU1FBMi84WEFBQUFQblZ5cDhlL0JCaFJRc0lnMnpqVHpqbnRZd0FpcFpqSDRMYlAyaXJLYXFzPSIsImF6cCI6Ijg3NmNjYTI3LTZlMjMtNDIxOC1iZTE4LTg3YTFkMmQ0ZjE5NSIsImF6cGFjciI6IjEiLCJyaCI6IjAuQVlJQUczdWNKd2E2ZTBLbWdjaWxTUzBwUFNmS2JJY2piaGhDdmhpSG9kTFU4WldDQUFBLiIsInRpZCI6IjI3OWM3YjFiLWJhMDYtNDI3Yi1hNjgxLWM4YTU0OTJkMjkzZCIsInV0aSI6InEwX0o0cDdKSlVDYnVoeF81WThBQUEiLCJ2ZXIiOiIyLjAifQ.NWHhvpClovU0PjfVVIVAMIR2lqSqnuyZGygnc7Q0fx1jOFF5zKFHnS8JHx53MA1tJmY2u88EhiVFL5u4Go-Q5qUNrLAVzF-chfUbhfBpdeiVR8jidopANitFAdX1PP06cXKSS4c9jQm1p8eQIrV2K2GZr4hGnjXALPAwi-ntbeeqkCq7EeKQru0PVwWYxUUIj5MBrMyengLB82McNugSuxhq9TU2o7ArffeeycvmjzigXdFX-LIFttRs4AVxGZM-iRwNs3ph6whIxN94nr_x2rjpkqwicnuRctibiMpzRxvO39ANJyf7nX6FWa5uQN5KGwitp7RyaS8aEAvR-GO4EQ"

type Server struct {
	mu          sync.Mutex
	BearerToken string
	pb.UnimplementedAddDebitAccountServiceServer
}

func NewServer() *Server {
	return &Server{
		BearerToken: bearerToken,
	}
}

func (s *Server) AddDebitAccount(ctx context.Context, req *pb.AddDebitAccountRequest) (*pb.AddDebitAccountResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logrus.Info("Starting AddDebitAccount process")

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	logrus.WithFields(logrus.Fields{
		"Accounts": req.Accounts,
	}).Info("Received request data")

	jsonData, err := json.Marshal(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal request")
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", "https://apitest.nibss-plc.com.ng/nibsspayplus/v2/Accounts", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.WithError(err).Error("Error creating request")
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+s.BearerToken)

	logrus.Infof("Request headers: %v", request.Header)
	resp, err := client.Do(request)
	if err != nil {
		logrus.WithError(err).Error("Error making request")
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read response body")
		return nil, err
	}
	bodyString := string(bodyBytes)
	logrus.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"response":    bodyString,
	}).Info("Raw response")

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    bodyString,
		}).Error("Received non-200 response")
		return nil, fmt.Errorf("error response from API: %d", resp.StatusCode)
	}

	var response pb.AddDebitAccountResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logrus.WithError(err).Error("Failed to decode response")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		"message": response.Message,
		"success": response.Success,
	}).Info("Parsed AddDebitAccount response")

	return &response, nil
}
