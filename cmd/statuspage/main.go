package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	PAGE_ID = "lzkfrqf402l6"
	API_KEY = os.Getenv("API_KEY")
)

type Status int8

const (
	Operational Status = iota + 1
	DegradedPerformance
	PartialOutage
	MajorOutage
	UnderMaintenance
)

func (status Status) String() string {
	switch status {
	case Operational:
		return "operational"
	case DegradedPerformance:
		return "degraded_performance"
	case PartialOutage:
		return "partial_outage"
	case MajorOutage:
		return "major_outage"
	case UnderMaintenance:
		return "under_maintenance"
	default:
		panic("unknown status")
	}
}

func main() {

	mercury := NewService(ComponentMercury, mercuryHealthCheck)
	go mercury.Run()

	btcMainnet := NewService(ComponentBtcMainnet, func() error {
		return addressBalanceCheck("/btc/mainnet","1D4NXvNvjucShZeyLsDzYz1ky2W8gYKQH7")
	})
	go btcMainnet.Run()

	btcTestnet := NewService(ComponentBtcTestnet, func() error {
		return addressBalanceCheck("/btc/testnet", "n4Vyt86t8bLyTogPBNHcP7qKgJbQHXjwTJ")
	})
	go btcTestnet.Run()

	zecMainnet := NewService(ComponentZecMainnet, func() error {
		return addressBalanceCheck("/zec/mainnet", "t1VvyYFo4iEQ3JChsHJ37go7gDghDTGVhnu")
	})
	go zecMainnet.Run()

	zecTestnet := NewService(ComponentZecTestnet, func() error {
		return addressBalanceCheck("/zec/testnet", "tmYsPB3SxYL6sRZhSYdYaJJSHnPguWPDQe2")
	})
	go zecTestnet.Run()

	bchMainnet := NewService(ComponentBchMainnet, func() error {
		return addressBalanceCheck("/bch/mainnet", "qzzyfwmnz3dlld7svwzn53xzr6ycz5kwavpd9uqf4l")
	})
	go bchMainnet.Run()

	bchTestnet := NewService(ComponentBchTestnet, func() error {
		return addressBalanceCheck("/bch/testnet", "qpn37uz8sqctxem3tfxayz09pr8w358hl5pvhd4twx")
	})
	go bchTestnet.Run()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<- done
}

type RequestUpdateComponent struct {
	Description        string `json:"description,omitempty"`
	Status             string `json:"status"`
	OnlyShowIfDegraded bool   `json:"only_show_if_degraded,omitempty"`
	Showcase           bool   `json:"showcase,omitempty"`
}

func UpdateStatusPage(componentID string, status Status) error {
	url := fmt.Sprintf("https://api.statuspage.io/v1/pages/%v/components/%v?api_key=%v", PAGE_ID, componentID, API_KEY)
	body := struct {
		Component RequestUpdateComponent `json:"component"`
	}{
		Component: RequestUpdateComponent{
			Description:        "",
			Status:             status.String(),
			OnlyShowIfDegraded: false,
			Showcase:           true,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)

	request, err:= http.NewRequest("PATCH", url, buf)
	if err != nil {
		return err
	}
	client := new(http.Client)
	response ,err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK{
		return fmt.Errorf("invalid status code, expected = 200, got = %v", response.StatusCode)
	}
	return nil
}
