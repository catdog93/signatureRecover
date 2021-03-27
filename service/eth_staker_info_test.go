package service

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common/math"
	signer "github.com/ethereum/go-ethereum/signer/core"
	"testing"
)

type want struct {
	isValid bool
	err     error
}

func TestIsValidSignature(t *testing.T) {
	var tests = []struct {
		signature string
		want      bool
	}{
		{"0x1379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc1b", true},
		{"01379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc1b", false},
		{"x1379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc1b", false},
		{"1379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc1b", true},
		{"", false},
	}

	for _, tt := range tests {
		actualResult := isValidSignature(tt.signature)
		if actualResult != tt.want {
			t.Errorf("got %v, want %v . Signature is %v", actualResult, tt.want, tt.signature)
		}
	}
}

func TestIsValidEmail(t *testing.T) {
	var tests = []struct {
		email string
		want  bool
	}{
		{"vanya234234", false},
		{"vanya234234@ever", false},
		{"vanya234234@everstake.one", true},
		//{"v#$%234@everstake.one", false},
		//{"$%234@everstake.one", false},
		{"VAnya@everstake.one", true},
		{"mail123@everstake.сom", false},
		{"vanya123@gmail.com", true},
		{"vanya123@@icloud.com", false},
		{"vanya123@icloud.com", true},
		{"yadya@yandex.com", true},
		{"v@.one", false},
		{"v@d.one", false},
		{"", false},
	}

	for _, tt := range tests {
		actualResult := isValidEmail(tt.email)
		if actualResult != tt.want {
			t.Errorf("got %v\n , want %v\n . Email is %v\n", actualResult, tt.want, tt.email)
		}
	}
}

func TestStakerInfoToTypedDataMessage(t *testing.T) {
	var tests = []struct {
		stakerInfo *smodels.EthStakerInfo
		want       *signer.TypedData
	}{
		{&smodels.EthStakerInfo{
			Address:  "0x73500c296d4863cdd01fb30c231db927d007f26c",
			Email:    "test@email.com",
			Telegram: "telegramHandle",
			ChainID:  3,
			Salt:     "qqqqqqqqqqqqqqqq",
		}, &signer.TypedData{
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
				ChainId: math.NewHexOrDecimal256(3),
				Version: EIP712DomainVersion,
				Salt:    "qqqqqqqqqqqqqqqq",
			},
			Message: signer.TypedDataMessage{
				"email":    "test@email.com",
				"telegram": "telegramHandle",
			},
		}},
	}

	for _, tt := range tests {
		actualResult := stakerInfoToTypedDataMessage(tt.stakerInfo)
		if actualResult.Domain.Name != tt.want.Domain.Name {
			t.Errorf("got %v ,\n want %v .\n", actualResult.Domain.Name, tt.want.Domain.Name)
		}
		if actualResult.Domain.Salt != tt.want.Domain.Salt {
			t.Errorf("got %v ,\n want %v .\n", actualResult.Domain.Salt, tt.want.Domain.Salt)
		}
		neg, value := actualResult.Domain.ChainId.MarshalText()
		neg2, value2 := tt.want.Domain.ChainId.MarshalText()
		if !bytes.Equal(neg, neg2) || value != value2 {
			t.Errorf("got %v ,\n want %v .\n", actualResult.Domain.ChainId, tt.want.Domain.ChainId)
		}
		if actualResult.Domain.Version != tt.want.Domain.Version {
			t.Errorf("got %v ,\n want %v .\n", actualResult.Domain.Version, tt.want.Domain.Version)
		}
		if actualResult.Message["email"] != tt.want.Message["email"] {
			t.Errorf("got %v, want %v . Email", actualResult.Message["email"], tt.want.Message["email"])
		}
		if actualResult.Message["telegram"] != tt.want.Message["telegram"] {
			t.Errorf("got %v, want %v . Telegram", actualResult.Message["telegram"], tt.want.Message["telegram"])
		}
		if actualResult.PrimaryType != tt.want.PrimaryType {
			t.Errorf("got %v ,\n want %v .\n", actualResult.PrimaryType, tt.want.PrimaryType)
		}
	}
}

func TestCheckEthStakerInfo(t *testing.T) {
	var tests = []struct {
		stakerInfo *smodels.EthStakerInfo
		want
	}{
		{&smodels.EthStakerInfo{
			Address:  "0x73500c296d4863cdd01fb30c231db927d007f26c",
			Email:    "test@email.com",
			Telegram: "telegramHandle",
			ChainID:  3,
			Salt:     "qqq",
			Signature: "0x1379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742" +
				"b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc" +
				"1b",
		}, want{true, nil}},
		{&smodels.EthStakerInfo{
			Address:  "0x73500c296d4863cdd01fb30c231db927d007f26c",
			Email:    "test@email.com",
			Telegram: "telegramHandle",
			ChainID:  3,
			Salt:     "qqq",
			Signature: "1379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742" +
				"b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc" +
				"1b",
		}, want{isValid: true, err: nil}},
		{&smodels.EthStakerInfo{
			Address:  "0x73500c296d4863cdd01fb30c231db927d007f26c",
			Email:    "test@email.com",
			Telegram: "telega",
			ChainID:  3,
			Salt:     "qqq",
			Signature: "1379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742" +
				"b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc" +
				"1b",
		}, want{isValid: false, err: nil}},
		{&smodels.EthStakerInfo{
			Address:  "0x73500c296d4863cdd01fb30c231db927d007f26c",
			Email:    "test@email.com",
			Telegram: "telegramHandle",
			ChainID:  3,
			Salt:     "qqq",
			Signature: "3379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742" +
				"b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc" +
				"1b",
		}, want{isValid: false, err: nil}},
	}

	for _, tt := range tests {
		actualResult, _ := checkEthStakerInfo(tt.stakerInfo)
		if actualResult != tt.want.isValid {
			t.Errorf("got %v, want %v . isValid", actualResult, tt.want)
		}
	}
}
