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
	Signature = "0xdf9b1e7ced2f7d2def2ed1fc7fd740d5bfb9bba99ad84252475f84b47ba480" +
		"da42864504cd61981b19b2393c764115eb7af1d6d61426937f4fe9baa8f6a50c70" +
		"1b"
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
				{Name: "verifyingContract", Type: "address"},
				{Name: "version", Type: "string"},
				{Name: "salt", Type: "string"},
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

	fmt.Println(hex.EncodeToString(domainSeparator)) // d886b5763ff830579c2de06a7b8ba98c3cb819202d5c6010d030227042f44603
	fmt.Println(hex.EncodeToString(typedDataHash))   // 4ec1628d891f1a5d124d4e9c9a2adea6cb3f11bd23ae4bf49a6c92b58035b12f

	fmt.Println(hex.EncodeToString(rawData)) // 1901d886b5763ff830579c2de06a7b8ba98c3cb819202d5c6010d030227042f446034ec1628d891f1a5d124d4e9c9a2adea6cb3f11bd23ae4bf49a6c92b58035b12f
	fmt.Println("1901237b71f9fd1f5397d70d9ce661789f677895efcf97cd66f0818786662959cd6e4ec1628d891f1a5d124d4e9c9a2adea6cb3f11bd23ae4bf49a6c92b58035b12f")

	fmt.Println(challengeHash) // 0x8f7d50ceb9b6b006a2f3d7a9293f2b02148f6f13731e8c922dfad3f6860e5682
	fmt.Println("0xb71caec68287c3a6805e093e9d7866f2a4830c47d73fe3cacba5e6fda2116b6b")
	sigBytes, err := hex.DecodeString(Signature)
	if err != nil {
		log.Println(err)
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
