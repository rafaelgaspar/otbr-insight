package service

import (
	"strings"

	"github.com/otbr-insight/otbr-insight/internal/model"
)

// enrichDeviceLinkMetrics copies link fields from the latest topology snapshot
// into the device inventory when the REST device collection omitted them. This
// is the normal path when live meshdiag is unavailable but OTBR network
// diagnostics still report childTable/children/routerNeighbors TLVs.
func enrichDeviceLinkMetrics(inventory *model.DeviceInventory, topology model.Topology) {
	if inventory == nil || len(inventory.Items) == 0 || len(topology.Links) == 0 {
		return
	}
	childLink := map[string]model.TopologyLink{}
	for _, link := range topology.Links {
		if link.Type != "child" {
			continue
		}
		key := strings.ToLower(link.Target)
		if key == "" {
			continue
		}
		childLink[key] = link
	}
	routerLink := map[string]model.TopologyLink{}
	for _, link := range topology.Links {
		if link.Type != "router" {
			continue
		}
		for _, endpoint := range []string{link.Source, link.Target} {
			endpoint = strings.ToLower(endpoint)
			if endpoint == "" {
				continue
			}
			existing, ok := routerLink[endpoint]
			if !ok || linkScore(link) > linkScore(existing) {
				routerLink[endpoint] = link
			}
		}
	}
	for i := range inventory.Items {
		device := &inventory.Items[i]
		if device.RSSI != nil || device.LinkMargin != nil || device.LinkQuality != nil {
			continue
		}
		id := deviceKey(device)
		if id == "" {
			continue
		}
		if link, ok := childLink[id]; ok {
			applyLink(device, link)
			continue
		}
		if link, ok := routerLink[id]; ok {
			applyLink(device, link)
		}
	}
}

func deviceKey(device model.Device) string {
	if device.ExtendedAddress != "" {
		return strings.ToLower(device.ExtendedAddress)
	}
	return strings.ToLower(device.ID)
}

func linkScore(link model.TopologyLink) int {
	score := 0
	if link.AverageRSSI != nil {
		score += 2
	}
	if link.LastRSSI != nil {
		score++
	}
	if link.LinkMargin != nil {
		score++
	}
	if link.LinkQuality != nil {
		score++
	}
	return score
}

func applyLink(device *model.Device, link model.TopologyLink) {
	device.RSSI = link.AverageRSSI
	if device.RSSI == nil {
		device.RSSI = link.LastRSSI
	}
	device.LinkMargin = link.LinkMargin
	if device.LinkQuality == nil {
		device.LinkQuality = link.LinkQuality
	}
}
