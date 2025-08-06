package generators

import (
	"fmt"
	"math/rand"
	"time"
)

// randomizers

func iso8601Now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000000Z")
}

func unixSeconds() int64 {
	return time.Now().Unix()
}

func randomALBType() string {
	types := []string{"http", "https", "h2"}
	return types[rand.Intn(len(types))]
}

func randomVPCAction() string {
	types := []string{"ACCEPT", "REJECT"}
	return types[rand.Intn(len(types))]
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
	)
}

func randomPort() int {
	return rand.Intn(65535-1024) + 1024
}

func randomAWSAccountID() string {
	// Generate a 12-digit number, ensuring the first digit is not 0
	accountID := rand.Intn(9) + 1
	for i := 1; i < 12; i++ {
		accountID = accountID*10 + rand.Intn(10)
	}
	return fmt.Sprintf("%012d", accountID)
}
