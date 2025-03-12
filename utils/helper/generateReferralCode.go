package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func GenerateReferralCode() string {
	randBytes := make([]byte, 2) 
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err)
	}

	randString := hex.EncodeToString(randBytes)

	uuidPart := uuid.New().String()[:4]

	return strings.ToUpper(fmt.Sprintf("%s-%s", uuidPart, randString))
}
