package service

import (
	"bitbucket.org/everstake/everstake-common/log"
	"bitbucket.org/everstake/everstake-dashboard/dmodels"
	"bitbucket.org/everstake/everstake-dashboard/filters"
	"bitbucket.org/everstake/everstake-dashboard/helpers"
	"bitbucket.org/everstake/everstake-dashboard/smodels"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/wedancedalot/decimal"
	"go.uber.org/zap"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	endpoint          = "api"
	logStakeDeposited = "TokensClaimed" // monitoring event, manual claim

	precision = 18
)

var precisionDiv = decimal.New(1, precision)

type (
	latestBlockNumberResponse struct {
		Result string `json:"result"`
	}
	ethLogsResponse struct {
		Status string `json:"status"`
		Result []struct {
			Data        string `json:"data"`
			BlockNumber string `json:"blockNumber"`
			Timestamp   string `json:"timeStamp"`
		} `json:"result"`
	}
	ethEventsStakeDeposited struct {
		ethABI      []byte
		blockNumber uint64
	}
)

// eventStakeDeposited represents StakeDeposited event payload
type eventStakeDeposited struct {
	Staker    common.Address
	Value     *big.Int
	Validator []byte
}

func (s *ServiceFacade) GetEthStakeDepositedEvents(request smodels.EthStakeDepositedEventsRequest) (resp []smodels.EthStakeDepositedEventsResponse, err error) {
	filter := filters.EthStakeDepositedEvents{
		Addresses: request.Addresses,
		Pagination: filters.Pagination{
			Page:  request.Page,
			Limit: request.Limit,
		},
	}
	items, err := s.dao.GetEthDepositEventsPaginated(filter)
	if err != nil {
		return nil, err
	}
	resp = make([]smodels.EthStakeDepositedEventsResponse, len(items))
	for index, item := range items {
		resp[index].Amount = item.Amount
		resp[index].Staker = item.Staker
		resp[index].Validator = item.Validator
	}
	return resp, nil
}

func (s *ServiceFacade) ValidateGetEthStakeDepositedEventsRequest(request smodels.EthStakeDepositedEventsRequest) (bool, string) {
	if len(request.Addresses) == 0 {
		return false, addressRequired
	}
	for _, address := range request.Addresses {
		valid := s.IsValidAddress(string(smodels.CurrencyCodeEthereum), address)
		if !valid {
			return false, fmt.Sprintf(invalidAddress, address)
		}
	}
	return true, ""
}

func (s *ServiceFacade) GetNewEthStakeDepositedEvents() {
	response, err := s.getLatestBlockNumber()
	if err != nil {
		log.Error("GetNewEthStakeDepositedEvents: service.getLatestBlockNumber", zap.Error(err))
		return
	}
	latestBlockNumber, err := strconv.ParseUint(response.Result, 0, 64)
	if err != nil {
		log.Error("GetNewEthStakeDepositedEvents: service.ParseUint latestBlockNumber", zap.Error(err))
		return
	}

	if s.ethEvents.blockNumber >= latestBlockNumber {
		return
	}

	resp, err := s.getLogs(s.ethEvents.blockNumber, latestBlockNumber)
	if err != nil {
		log.Error("GetNewEthStakeDepositedEvents: service.getLogs", zap.Error(err))
		return
	}

	s.ethEvents.blockNumber = latestBlockNumber

	event := eventStakeDeposited{}
	for _, result := range resp.Result {
		timestamp, err := strconv.ParseInt(result.Timestamp, 0, 64)
		if err != nil {
			log.Error("GetNewEthStakeDepositedEvents: service.ParseInt Timestamp", zap.Error(err))
			return
		}
		blockNumber, err := strconv.ParseUint(result.BlockNumber, 0, 64)
		if err != nil {
			log.Error("GetNewEthStakeDepositedEvents: service.ParseUInt BlockNumber", zap.Error(err))
			return
		}

		dataBytes, err := hex.DecodeString(result.Data[2:])
		if err != nil {
			log.Error("GetNewEthStakeDepositedEvents: service.DecodeString", zap.Error(err))
			return
		}
		err = event.Unmarshal(dataBytes, s.ethEvents.ethABI)
		if err != nil {
			log.Error("GetNewEthStakeDepositedEvents: service.Unmarshal dataBytes into struct", zap.Error(err))
			return
		}

		amount := decimal.NewFromBigInt(event.Value, 0).Div(precisionDiv)

		err = s.dao.CreateEthDepositEvent(dmodels.EthStakeDepositedEvent{
			Staker:      event.Staker.String(),
			Amount:      amount,
			Validator:   "0x" + hex.EncodeToString(event.Validator),
			BlockNumber: blockNumber,
			CreatedAt:   time.Unix(timestamp, 0),
		})
		if err != nil {
			log.Error("GetNewEthStakeDepositedEvents: dao.CreateEthDepositEvent", zap.Error(err))
		}
	}
}

func (s *ServiceFacade) getLatestBlockNumber() (response latestBlockNumberResponse, err error) {
	url := fmt.Sprintf("%s/%s", s.cfg.EthereumEvents.APIURL, endpoint)
	params := map[string]string{
		"module": "proxy",
		"action": "eth_blockNumber",
		"apikey": s.cfg.EthereumEvents.APIKey,
	}
	err = helpers.HTTPGet(url, params, &response)
	return response, err
}

func (s *ServiceFacade) getLogs(fromBlock, toBlock uint64) (response ethLogsResponse, err error) {
	url := fmt.Sprintf("%s/%s", s.cfg.EthereumEvents.APIURL, endpoint)
	params := map[string]string{
		"module":    "logs",
		"action":    "getLogs",
		"fromBlock": fmt.Sprint(fromBlock),
		"toBlock":   fmt.Sprint(toBlock),
		"address":   s.cfg.EthereumEvents.PoolAddress,
		"topic0":    s.cfg.EthereumEvents.TopicId,
		"apikey":    s.cfg.EthereumEvents.APIKey,
	}
	err = helpers.HTTPGet(url, params, &response)
	return response, err
}

// Unmarshal data into event structure
func (e *eventStakeDeposited) Unmarshal(data, abiBytes []byte) error {
	event, err := getEventStakeDeposited(abiBytes)
	if err != nil {
		return err
	}
	err = unmarshalEvent(e, data, event)
	if err != nil {
		return err
	}
	return err
}

func getEventStakeDeposited(abiBytes []byte) (event abi.Event, err error) {
	poolStakeDeposited, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return event, err
	}
	event, ok := poolStakeDeposited.Events[logStakeDeposited]
	if !ok {
		return event, fmt.Errorf("eth Events not found %s in events map", logStakeDeposited)
	}
	return event, nil
}

// unmarshalEvent blockchain log into the event structure
// `dest` must be a pointer to initialized structure
func unmarshalEvent(dest interface{}, data []byte, e abi.Event) error {
	a := abi.ABI{Events: map[string]abi.Event{"e": e}}
	return a.Unpack(dest, "e", data)
}
