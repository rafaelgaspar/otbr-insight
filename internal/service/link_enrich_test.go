package service

import (
	"testing"

	"github.com/otbr-insight/otbr-insight/internal/model"
)

func TestEnrichDeviceLinkMetricsFromChildLink(t *testing.T) {
	rssi := -64
	margin := 12
	inventory := &model.DeviceInventory{
		Items: []model.Device{{ID: "0102030405060708", ExtendedAddress: "0102030405060708"}},
	}
	topology := model.Topology{
		Links: []model.TopologyLink{{
			Type: "child", Target: "0102030405060708",
			AverageRSSI: &rssi, LinkMargin: &margin,
		}},
	}
	enrichDeviceLinkMetrics(inventory, topology)
	if inventory.Items[0].RSSI == nil || *inventory.Items[0].RSSI != -64 {
		t.Fatalf("RSSI = %v, want -64", inventory.Items[0].RSSI)
	}
	if inventory.Items[0].LinkMargin == nil || *inventory.Items[0].LinkMargin != 12 {
		t.Fatalf("LinkMargin = %v, want 12", inventory.Items[0].LinkMargin)
	}
}

func TestEnrichDeviceLinkMetricsFromRouterLink(t *testing.T) {
	rssi := -72
	margin := 28
	inventory := &model.DeviceInventory{
		Items: []model.Device{{ID: "aaaabbbbccccdddd", ExtendedAddress: "aaaabbbbccccdddd", Role: "router"}},
	}
	topology := model.Topology{
		Links: []model.TopologyLink{{
			Type: "router", Source: "1111222233334444", Target: "aaaabbbbccccdddd",
			AverageRSSI: &rssi, LinkMargin: &margin,
		}},
	}
	enrichDeviceLinkMetrics(inventory, topology)
	if inventory.Items[0].RSSI == nil || *inventory.Items[0].RSSI != -72 {
		t.Fatalf("RSSI = %v, want -72", inventory.Items[0].RSSI)
	}
}
