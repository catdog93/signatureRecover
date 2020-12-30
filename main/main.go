package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signer "github.com/ethereum/go-ethereum/signer/core"
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
			"ContactInfo": []signer.Type{
				{Name: "email", Type: "string"},
				{Name: "telegram", Type: "string"},
			},
			"EIP712Domain": []signer.Type{
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "version", Type: "string"},
				{Name: "verifyingContract", Type: "address"},
				{Name: "salt", Type: "bytes32"},
			},
		},
		PrimaryType: "ContactInfo",
		Domain: signer.TypedDataDomain{
			Name:              "Everstake ETH2 Staker",
			Version:           "1",
			VerifyingContract: "0x4aefd9A9BF4d0F19CD217ddB6467F3ad1e0A21FC",
			Salt:              "0x1122334455667788990011223344556677889900112233445566778899001122",
			ChainId:           math.NewHexOrDecimal256(3),
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

	fmt.Println(challengeHash) // 0x225b900e8ed24c6f9910df7d36e6da93e8d1ad7894514df73df1cf07f0053872
	fmt.Println("0x" + "5f8e30e1754bb3b1932caed72165313bea2b3e012d9f9eb948815714d63ff8e1")
}
