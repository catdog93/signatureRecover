package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
	"log"
)

const (
	Address = "0x73500c296d4863cdd01fb30c231db927d007f26c"
)

var Signature = []byte("0x1379af5b868812d055aafd905693990c6103ba5ecfd524249b7048d5d5c0742" +
	"b42d4e3e7aca742f09acb881ae34658c5f399f9b8c5d29858000852e2ee02bdcc" +
	"1b")

var (
	signerData = signer.TypedData{
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
			Name:    "Everstake ETH2 Staker",
			ChainId: math.NewHexOrDecimal256(3),
			Version: "1",
			Salt:    "qqq",
		},
		Message: signer.TypedDataMessage{
			"email":    "test@email.com",
			"telegram": "telegramHandle",
		},
	}
)

func main() {
	typedDataHash, _ := signerData.HashStruct(signerData.PrimaryType, signerData.Message)
	domainSeparator, _ := signerData.HashStruct("EIP712Domain", signerData.Domain.Map())

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	challengeHash := crypto.Keccak256Hash(rawData)

	_, err := crypto.SigToPub(challengeHash.Bytes(), Signature)
	if err != nil {
		log.Fatal(err)
	}

	_, err = crypto.Ecrecover(challengeHash.Bytes(), Signature)
	if err != nil {
		log.Fatal(err)
	}
}
