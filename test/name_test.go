package ubloxbluetooth

import (
	"fmt"
	"testing"
)

func TestStringToASCII(t *testing.T) {
	sample := "hello"
	for i := 0; i < len(sample); i++ {
		fmt.Printf("%x", sample[i])
	}
	fmt.Printf("\n%x\n", sample)
}

func TestNameFunctions(t *testing.T) {
	ub, err := setupBluetooth()
	if err != nil {
		t.Fatalf("setupBluetooth error %v\n", err)
	}
	defer ub.Close()

	err = connectToDevice("FF716C704ECBr", func(t *testing.T) error {
		err := ub.WriteName("TestName")
		if err != nil {
			t.Errorf("WriteName error %v\n", err)
		}

		name, err := ub.ReadName()
		if err != nil {
			t.Errorf("WriteName error %v\n", err)
		}

		fmt.Printf("Device Name %s\n", name)
		return nil
	}, ub, t)
	if err != nil {
		t.Errorf("TestReboot error %v\n", err)
	}
}
