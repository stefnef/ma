package model

import (
	"log"
	"testing"
	"time"
)

func TestDemoThreeBookings(t *testing.T) {
	if _, err := testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	log.Println("*******************************************************************")
	log.Println("************************** Booking test ***************************")
	log.Println("Scenario: a client books 3 flights at bonus level 'low'")
	log.Println("*******************************************************************")

	client := NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}

	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		if err := client.Booking(1, utLowLevelID); err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	time.Sleep(time.Second)
	log.Println()
	log.Println("Received bonus codes: ")
	for _, code := range client.BonusCodes {
		log.Printf("- %s (valid for bonus level %s)\n", (*code).CodeID, (*code).ValidFor.BonusID)
	}
	log.Println()
}

func TestDemoAccess(t *testing.T) {
	if _, err := testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	log.Println("********************************************************************")
	log.Println("*************************** Access test ****************************")
	log.Println("Scenario: a client books 3 flights at bonus level 'middle' and requests access")
	log.Println("********************************************************************")

	client := NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	log.Println()
	log.Println("The client books three times and receives bonus codes.")
	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		if err := client.Booking(1, utMiddleLevelID); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	time.Sleep(time.Second)
	log.Println()
	log.Println("Received bonus codes: ")
	for _, code := range client.BonusCodes {
		log.Printf("- %s (valid for bonus level %s)\n", (*code).CodeID, (*code).ValidFor.BonusID)
	}

	log.Println()
	log.Println("The client requests access.")
	time.Sleep(time.Second)
	if err := client.AccessBonusSystem(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	time.Sleep(time.Second)
	log.Println("Access was granted for level:")
	for level, token := range client.BLevelToTokens {
		log.Printf(" - %s (relation token: %s, recovery token: %s)\n", level, token, client.BLevelToRecovery[level])
	}
	log.Println()
}

func TestDemoParticipate(t *testing.T) {
	if _, err := testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	log.Println("*********************************************************************")
	log.Println("************************ Participation test *************************")
	log.Println("Scenario: a client gets access for the levels low and middle.\n" +
		"The client tries to participate at level 'low', 'middle' and 'high'.\n" +
		"Only participation at level 'high' will fail.")
	log.Println("*********************************************************************")

	client := NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	time.Sleep(time.Second)
	log.Println()
	log.Println("The client books three times and receives bonus codes.")
	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		if err := client.Booking(1, utMiddleLevelID); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	time.Sleep(time.Second)
	log.Println()
	log.Println("Received bonus codes: ")
	for _, code := range client.BonusCodes {
		log.Printf("- %s (valid for bonus level %s)\n", (*code).CodeID, (*code).ValidFor.BonusID)
	}

	time.Sleep(time.Second)
	log.Println()
	log.Println("The client requests access.")
	time.Sleep(time.Second)
	if err := client.AccessBonusSystem(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	time.Sleep(time.Second)
	log.Println("Access was granted for level:")
	for level, token := range client.BLevelToTokens {
		log.Printf(" - %s (relation token: %s, recovery token: %s)\n", level, token, client.BLevelToRecovery[level])
	}

	time.Sleep(time.Second)
	log.Println()
	log.Println("The client participates for level low:")
	time.Sleep(time.Second)
	if bData, err := client.Participate(utLowLevelID); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		time.Sleep(time.Second)
		log.Printf("Received bonus data: %s\n", bData)
	}

	log.Println()
	log.Println("The client participates for level middle:")
	time.Sleep(time.Second)
	if bData, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		time.Sleep(time.Second)
		log.Printf("Received bonus data: %s\n", bData)
	}

	log.Println()
	log.Println("The relation and recovery tokens changed for both bonus levels:")
	for level, token := range client.BLevelToTokens {
		log.Printf(" - %s (relation token: %s, recovery token: %s)\n", level, token, client.BLevelToRecovery[level])
	}

	log.Println()
	log.Println("Participation for bonus level 'high' is not possible because this level was not accessed:")
	time.Sleep(time.Second)
	_, err := client.Participate(utHighLevelID)
	time.Sleep(time.Second)
	log.Printf("Error message: %s", err)

	log.Println()
}

func TestDemoRestoration(t *testing.T) {
	if _, err := testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	log.Println("*********************************************************************")
	log.Println("************************* Restoration test **************************")
	log.Println("Scenario: a client gets access for the levels middle and participates\n" +
		"two times. The client restores the access data after that.")
	log.Println("*******************************************************************")

	client := NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	log.Println()
	log.Println("The client books three times and receives bonus codes.")
	time.Sleep(time.Second)
	for i := 0; i < 3; i++ {
		if err := client.Booking(1, utMiddleLevelID); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	time.Sleep(time.Second)
	log.Println()
	log.Println("The client requests access.")
	time.Sleep(time.Second)
	if err := client.AccessBonusSystem(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	time.Sleep(time.Second)
	log.Println("Access granted for level:")
	for level, token := range client.BLevelToTokens {
		log.Printf(" - %s (relation token: %s, recovery token: %s)\n", level, token, client.BLevelToRecovery[level])
	}

	log.Println()
	log.Println("The client participates two times for level middle.")
	log.Println("1st participation:")
	time.Sleep(time.Second)
	if bData, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		time.Sleep(time.Second)
		log.Printf("Received bonus data of 1st participation: %s\n", bData)
	}

	log.Println("2nd participation:")
	time.Sleep(time.Second)
	if bData, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		time.Sleep(time.Second)
		log.Printf("Received bonus data of 2nd participation: %s\n", bData)
	}

	log.Println("The relation and recovery tokens only changed for bonus level middle:")
	for level, token := range client.BLevelToTokens {
		log.Printf(" - %s (relation token: %s, recovery token: %s)\n", level, token, client.BLevelToRecovery[level])
	}

	log.Println()
	log.Println("The client restores the access data:")
	time.Sleep(time.Second)
	client = NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	time.Sleep(time.Second)
	if bData, err := client.RestoreFromMnemonic(1, utMnemonic, 2); err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		time.Sleep(time.Second)
		log.Println("The restoration was successful.")
		log.Printf("The client received the bonus data of the 2nd participation: %s", bData[utMiddleLevelID])
	}

	log.Println()
}
