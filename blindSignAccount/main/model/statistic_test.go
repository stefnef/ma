package model

import (
	"testing"
)

func TestNewStatisticArray(t *testing.T) {
	bLevelLow := setupBonusLevel()
	if bLevelLow == nil {
		t.Error("bLevel is nil")
		t.Fail()
	}

	if bLevelLow != nil && bLevelLow.ActionVariants != nil {
		for _, actionVariant := range bLevelLow.ActionVariants {
			for i := 0; i < 9; i++ {
				if actionVariant.Statistic[i] == nil {
					t.Error("statistic is a nil entry")
					t.Fail()
				}
			}
		}
	}
}

func TestSaveReadWrite(t *testing.T) {
	bLevelLow := setupBonusLevel()
	variant := bLevelLow.ActionVariants[ActionBooking]

	variant.SeedToAddress["seed"] = "adr1"
	SaveWrite(StatSeedToAddress, variant)
	if variant.Statistic[StatSeedToAddress].NrReads != 0 ||
		variant.Statistic[StatSeedToAddress].NrWrites != 1 ||
		variant.Statistic[StatSeedToAddress].Length != 1 {
		t.Errorf("wrong nr reads %d or writes %d or length %d", variant.Statistic[StatSeedToAddress].NrReads,
			variant.Statistic[StatSeedToAddress].NrWrites, variant.Statistic[StatSeedToAddress].Length)
		t.Fail()
	}

	variant.SeedToAddress["seed"] = "adr2"
	SaveWrite(StatSeedToAddress, variant)
	if variant.Statistic[StatSeedToAddress].NrReads != 0 ||
		variant.Statistic[StatSeedToAddress].NrWrites != 2 ||
		variant.Statistic[StatSeedToAddress].Length != 1 {
		t.Errorf("wrong nr reads %d or writes %d or length %d", variant.Statistic[StatSeedToAddress].NrReads,
			variant.Statistic[StatSeedToAddress].NrWrites, variant.Statistic[StatSeedToAddress].Length)
		t.Fail()
	}

	variant.SeedToAddress["seed2"] = "adr3"
	SaveWrite(StatSeedToAddress, variant)
	if variant.Statistic[StatSeedToAddress].NrReads != 0 ||
		variant.Statistic[StatSeedToAddress].NrWrites != 3 ||
		variant.Statistic[StatSeedToAddress].Length != 2 {
		t.Errorf("wrong nr reads %d or writes %d or length %d", variant.Statistic[StatSeedToAddress].NrReads,
			variant.Statistic[StatSeedToAddress].NrWrites, variant.Statistic[StatSeedToAddress].Length)
		t.Fail()
	}

	SaveRead(StatSeedToAddress, variant)
	if variant.Statistic[StatSeedToAddress].NrReads != 1 ||
		variant.Statistic[StatSeedToAddress].NrWrites != 3 ||
		variant.Statistic[StatSeedToAddress].Length != 2 {
		t.Errorf("wrong nr reads %d or writes %d or length %d", variant.Statistic[StatSeedToAddress].NrReads,
			variant.Statistic[StatSeedToAddress].NrWrites, variant.Statistic[StatSeedToAddress].Length)
		t.Fail()
	}

	SaveRead(StatSeedToAccountID, variant)
	if variant.Statistic[StatSeedToAddress].NrReads != 1 ||
		variant.Statistic[StatSeedToAddress].NrWrites != 3 ||
		variant.Statistic[StatSeedToAddress].Length != 2 {
		t.Errorf("wrong nr reads %d or writes %d or length %d", variant.Statistic[StatSeedToAddress].NrReads,
			variant.Statistic[StatSeedToAddress].NrWrites, variant.Statistic[StatSeedToAddress].Length)
		t.Fail()
	}
	if variant.Statistic[StatSeedToAccountID].NrReads != 1 ||
		variant.Statistic[StatSeedToAccountID].NrWrites != 0 ||
		variant.Statistic[StatSeedToAccountID].Length != 0 {
		t.Errorf("wrong nr reads %d or writes %d or length %d", variant.Statistic[StatSeedToAccountID].NrReads,
			variant.Statistic[StatSeedToAccountID].NrWrites, variant.Statistic[StatSeedToAccountID].Length)
		t.Fail()
	}
}
