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
	types := []string{"http", "https"}
	return types[rand.Intn(len(types))]
}

func randomVPCAction() string {
	types := []string{"ACCEPT", "REJECT"}
	return types[rand.Intn(len(types))]
}

// randomIP from 72.16.101.0/24
func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		172,
		16,
		101,
		rand.Intn(256),
	)
}

// randomPort in range 58080 to 62000
func randomPort() int {
	return rand.Intn(62000-58080) + 58080
}

func randomAWSAccountID() string {
	// Generate a 12-digit number, ensuring the first digit is not 0
	accountID := rand.Intn(9) + 1
	for i := 1; i < 12; i++ {
		accountID = accountID*10 + rand.Intn(10)
	}
	return fmt.Sprintf("%012d", accountID)
}
