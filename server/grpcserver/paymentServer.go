package grpcserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "grpcTestConnection/server/payment/grpcserver/payment"
	"mime/multipart"

	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var keys = "8AHDle5EU9sxEmYUSQ2NwBEMVzlmrip2"

type Server struct {
	mu          sync.Mutex
	BearerToken string
	pb.UnimplementedInternalsServiceServer
	authMutex sync.Mutex
}

func NewServer() *Server {
	return &Server{}
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

func (s *Server) GetBankList(ctx context.Context, _ *emptypb.Empty) (*pb.GetBankResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logrus.Info("Fetching all bank list")

	url := "https://apitest.nibss-plc.com.ng/nibsspayplus/v2/Banks"
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logrus.WithError(err).Error("Error creating request")
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+s.BearerToken)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logrus.WithError(err).Error("Failed to send request")
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

	var response pb.GetBankResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logrus.WithError(err).Error("Failed to decode response")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		"data": response.Data,
	}).Info("Parsed bank list response")

	return &response, nil
}

func (s *Server) Authenticate(ctx context.Context, req *pb.AuthenticationRequest) (*pb.AuthenticationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logrus.Info("Starting authentication process")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("client_id", req.ClientId)
	w.WriteField("client_secret", req.ClientSecret)
	w.WriteField("scope", req.Scope)
	w.WriteField("grant_type", req.GrantType)
	w.Close()

	request, err := http.NewRequestWithContext(ctx, "POST", "https://apitest.nibss-plc.com.ng/v2/reset", &b)
	if err != nil {
		logrus.Errorf("Error creating request: %v", err)
		return nil, err
	}
	request.Header.Set("Content-Type", w.FormDataContentType())
	request.Header.Set("apikey", keys)

	logrus.Infof("Request headers: %v", request.Header)
	resp, err := client.Do(request)
	if err != nil {
		logrus.Errorf("Error making request: %v", err)
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

	var authResp pb.AuthenticationResponse
	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		logrus.WithError(err).Error("Failed to decode response")
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		"token_type":   authResp.TokenType,
		"access_token": authResp.AccessToken,
	}).Info("Parsed authentication response")
	if authResp.AccessToken == "" || authResp.TokenType == "" {
		logrus.Warn("Access token or token type is empty")
	}
	s.authMutex.Lock()
	s.BearerToken = authResp.AccessToken
	s.authMutex.Unlock()
	return &authResp, nil
}
