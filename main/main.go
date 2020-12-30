package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
	"log"
)

const (
	Signature = "3c99d3de749ced00853bf47eb8e408f192310b03ab755cc5aa6456c5507baa88" +
		"5bed78ebc1e802de48a9b927cefa01afbcdac74b65a35913cbb4c5bf3071a0ce" +
		"1c"
	Address = "0x73500c296d4863cdd01fb30c231db927d007f26c"
)

var (
	signerData = signer.TypedData{
		Types: signer.Types{
			"Challenge": []signer.Type{
				{Name: "address", Type: "address"},
				{Name: "nonce", Type: "string"},
				{Name: "timestamp", Type: "string"},
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
			Version: "1",
			Salt:    "0x1122334455667788990011223344556677889900112233445566778899001122",
			ChainId: math.NewHexOrDecimal256(3),
		},
		Message: signer.TypedDataMessage{
			"email":    "test@email.com",
			"telegram": "telegramHandle",
		},
	}
)

func main() {
	//hash := crypto.Keccak256Hash(data)
	typedDataHash, _ := signerData.HashStruct(signerData.PrimaryType, signerData.Message)
	domainSeparator, _ := signerData.HashStruct("EIP712Domain", signerData.Domain.Map())

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	challengeHash := crypto.Keccak256Hash(rawData)

	sigBytes, err := hex.DecodeString(Signature)
	if err != nil {
		log.Print(err)
	}

	sigBytes[64] -= 27

	sigPublicKey, err := crypto.SigToPub(challengeHash.Bytes(), sigBytes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sigPublicKey)

	sigAddress := crypto.PubkeyToAddress(*sigPublicKey)

	fmt.Println(sigAddress.String())
	fmt.Println("expected address:", Address)
}