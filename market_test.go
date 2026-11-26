package main

import "testing"

func TestPackSizesAndContents(t *testing.T) {
	for kind := range packKinds {
		for size := range packSizes {
			m := newModel("pack")
			m.money = 30
			m.packKind = kind
			m.packSize = size
			m.openPack()
			want := 3
			if size > 0 {
				want = 5
			}
			if len(m.packChoices) != want {
				t.Fatal(kind, size, len(m.packChoices))
			}
			m.pickPack("1")
			if size == 2 {
				if len(m.packChoices) != 4 {
					t.Fatal("mega closed early")
				}
				m.pickPack("1")
			}
			if len(m.packChoices) != 0 {
				t.Fatal("extra picks")
			}
			if !m.packSold {
				t.Fatal("pack still available")
			}
		}
	}
}
func TestUpgradeEligibility(t *testing.T) {
	m := newModel("shop")
	m.unlockRound = 3
	m.money = 10000
	for range 32 {
		m.openShop()
		if m.upgradeOffer < 0 {
			t.Fatal("missing upgrade")
		}
		m.buyUpgrade()
	}
	m.openShop()
	if m.upgradeOffer != -1 {
		t.Fatal("third tier offered")
	}
	m.startRound()
	if len(m.hand) != 10 || m.plays != 6 || m.discards != 5 || m.modSlots() != 7 || m.bagSlots() != 4 {
		t.Fatal("upgrade effects missing")
	}
}
func TestQueuedReroll(t *testing.T) {
	m := newModel("queued")
	m.openShop()
	m.money = 0
	for i, v := range tags {
		if v.Name == "Toll slip" {
			m.tagQueue = []int{i, i}
		}
	}
	m.updateShop("r")
	m.updateShop("r")
	if m.money != 0 || len(m.tagQueue) != 0 || m.freeRerolls != 0 || m.rerollCharge() != 4 {
		t.Fatal("free rerolls did not apply once each")
	}
}
