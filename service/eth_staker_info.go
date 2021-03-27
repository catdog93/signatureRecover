package service

import (
	"bitbucket.org/everstake/everstake-common/log"
	"bitbucket.org/everstake/everstake-dashboard/dmodels"
	"bitbucket.org/everstake/everstake-dashboard/smodels"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
	"go.uber.org/zap"
	"net"
	"regexp"
	"strings"
	"time"
)

const (
	emailRegex = "(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])"

	EIP712PrimaryType   = "EIP712Domain"
	EIP712DomainVersion = "1"
	EIP712DomainName    = "Everstake ETH2 Staker"

	invalidAddress   = "invalid address %s"
	addressRequired  = "address required"
	invalidSignature = "invalid signature"
	invalidEmail     = "invalid email"
	invalidTelegram  = "invalid telegram username"
	invalidChainID   = "invalid chainID"
	invalidSalt      = "invalid salt"
)

func (s *ServiceFacade) CheckEthStakerInfo(stakerInfo *smodels.EthStakerInfo) (bool, error) {
	return checkEthStakerInfo(stakerInfo)
}

func (s *ServiceFacade) FindEthStakerInfo(address string) (bool, error) {
	return s.dao.FindEthStakerInfo(address)
}

func (s *ServiceFacade) CreateEthStakerInfo(stakerInfo *smodels.EthStakerInfo) error {
	err := s.dao.CreateEthStakerInfo(dmodels.EthStakerInfo{
		Address:   stakerInfo.Address,
		Email:     stakerInfo.Email,
		Telegram:  stakerInfo.Telegram,
		Salt:      stakerInfo.Salt,
		ChainID:   stakerInfo.ChainID,
		Signature: stakerInfo.Signature,
	})
	return err
}

func (s *ServiceFacade) UpdateEthStakerInfo(stakerInfo *smodels.EthStakerInfo) error {
	err := s.dao.UpdateEthStakerInfo(dmodels.EthStakerInfo{
		Address:   stakerInfo.Address,
		Email:     stakerInfo.Email,
		Telegram:  stakerInfo.Telegram,
		Salt:      stakerInfo.Salt,
		Signature: stakerInfo.Signature,
		CreatedAt: time.Now(),
	})
	return err
}

func (s *ServiceFacade) ValidateEthStakerInfo(stakerInfo smodels.EthStakerInfo) (bool, string) {
	valid := s.IsValidAddress(string(smodels.CurrencyCodeEthereum), stakerInfo.Address)
	if !valid {
		return false, fmt.Sprintf(invalidAddress, stakerInfo.Address)
	}
	if !isValidEmail(stakerInfo.Email) {
		return false, invalidEmail
	}
	if !isValidTelegramUsername(stakerInfo.Telegram) {
		return false, invalidTelegram
	}
	if !isValidChainID(stakerInfo.ChainID) {
		return false, invalidChainID
	}
	if !isValidSalt(stakerInfo.Salt) {
		return false, invalidSalt
	}
	if !isValidSignature(stakerInfo.Signature) {
		return false, invalidSignature
	}
	return true, ""
}

func isValidEmail(email string) bool {
	r := regexp.MustCompile(emailRegex)
	if !r.MatchString(email) {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	mxs, err := net.LookupMX(parts[1])
	if err != nil || len(mxs) == 0 {
		return false
	}
	return true
}

func isValidTelegramUsername(username string) bool {
	return len(username) >= 5
}

func isValidSalt(salt string) bool {
	return len(salt) == 16
}

func isValidChainID(chainID int64) bool {
	return chainID >= 1 && chainID <= 5
}

func isValidSignature(signature string) bool {
	return len(signature) == 130 || (len(signature) == 132 && strings.HasPrefix(signature, "0x"))
}

func stakerInfoToTypedDataMessage(stakerInfo *smodels.EthStakerInfo) *signer.TypedData {
	return &signer.TypedData{
		Types: signer.Types{
			"ContactInfo": []signer.Type{
				{Name: "email", Type: "string"},
				{Name: "telegram", Type: "string"},
			},
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "version", Type: "string"},
				{Name: "salt", Type: "string"},
			},
		},
		PrimaryType: "ContactInfo",
		Domain: signer.TypedDataDomain{
			Name:    EIP712DomainName,
			ChainId: math.NewHexOrDecimal256(stakerInfo.ChainID),
			Version: EIP712DomainVersion,
			Salt:    stakerInfo.Salt,
		},
		Message: signer.TypedDataMessage{
			"email":    stakerInfo.Email,
			"telegram": stakerInfo.Telegram,
		},
	}
}

func messageToHash(data *signer.TypedData) ([]byte, error) {
	typedDataHash, err := data.HashStruct(data.PrimaryType, data.Message)
	if err != nil {
		return nil, err
	}
	domainSeparator, err := data.HashStruct(EIP712PrimaryType, data.Domain.Map())
	if err != nil {
		return nil, err
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	b := crypto.Keccak256Hash(rawData).Bytes()
	return b, err
}

func decodeMetamaskSignature(signature string) (sig []byte, err error) {
	if len(signature) == 132 && strings.HasPrefix(signature, "0x") {
		signature = signature[2:]
	}
	sig, err = hex.DecodeString(signature)
	if err != nil {
		return
	}
	if len(sig) != 65 {
		return nil, fmt.Errorf("invalid signature length: %d", len(sig))
	}
	if sig[64] != 27 && sig[64] != 28 {
		return nil, fmt.Errorf("invalid recovery id: %d", sig[64])
	}
	sig[64] -= 27
	return
}

func checkEthStakerInfo(stakerInfo *smodels.EthStakerInfo) (bool, error) {
	typedDataMessage := stakerInfoToTypedDataMessage(stakerInfo)
	challengeHash, err := messageToHash(typedDataMessage)
	if err != nil {
		return false, err
	}
	sig, err := decodeMetamaskSignature(stakerInfo.Signature)
	if err != nil {
		return false, err
	}
	pub, err := crypto.SigToPub(challengeHash, sig)
	if err != nil {
		return false, err
	}
	recoveredAddress := crypto.PubkeyToAddress(*pub)
	recoveredAddressLower := strings.ToLower(recoveredAddress.String())
	addressLower := strings.ToLower(stakerInfo.Address)
	if addressLower != recoveredAddressLower {
		log.Info("addresses don't match", zap.String("address", addressLower), zap.String("recoveredAddress", recoveredAddressLower))
		return false, fmt.Errorf("addresses don't match")
	}
	return true, nil
}
