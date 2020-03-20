package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	URL                 = "http://aws-lb-mercury-1187443577.us-east-1.elb.amazonaws.com"
	ComponentMercury    = "knx5j6jqnj4m"
	ComponentBtcMainnet = "gt7wggppkbz4"
	ComponentBtcTestnet = "cv00yh62ygyc"
	ComponentZecMainnet = "9jp5x719zygv"
	ComponentZecTestnet = "39dt0h6m6x89"
	ComponentBchMainnet = "n8r9thyjqkbr"
	ComponentBchTestnet = "jslmxxp7sbrg"
)

type Service struct {
	componentID string
	status      Status
	healthCheck func()error
}

func NewService( componentID string, healthCheck func()error) Service{
	return Service{
		componentID: componentID,
		status:      0,
		healthCheck: healthCheck,
	}
}

func (service *Service) Run() {
	for {
		if err := service.healthCheck(); err != nil {
			log.Printf("health check failed for component = %v, err = %v",service.componentID, err)
			service.updateStatus(MajorOutage)
			time.Sleep(time.Minute)
			continue
		}
		service.updateStatus(Operational)
		time.Sleep(30 * time.Second)
	}
}

func (service *Service) updateStatus(status Status) {
	if service.status == 0 || service.status != status{
		if err := UpdateStatusPage(service.componentID, status); err != nil {
			log.Printf("cannot update service status, err = %v", err)
			return
		}
		service.status = status
	}
}

func mercuryHealthCheck() error {
	response, err := http.Get(URL + "/health")
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("invalid status code, expected = 200, got = %v", response.StatusCode)
	}
	return nil
}

func addressBalanceCheck(postfix, address string) error{
	body := fmt.Sprintf("{\"jsonrpc\": \"1.0\", \"id\": 123, \"method\": \"listunspent\", \"params\": [ 6, 9999999, [ \"%v\" ]]}", address)
	buf := bytes.NewBuffer([]byte(body))
	response, err := http.Post(fmt.Sprintf("%v%v", URL, postfix), "application/json",buf )
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK{
		return fmt.Errorf("invalid status code, expected = 200, got = %v", response.StatusCode)
	}
	return nil
}
