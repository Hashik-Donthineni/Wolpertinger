package main

const (
	MaxReturnedBridges = 3
)

// GetBridges queries our backend to find and return bridges that we want
// tested by censorship measurement platforms like OONI.
func GetBridges(req *BridgeRequest) (*BridgeResponse, error) {

	i := 0
	resp := BridgeResponse{}

	bridges.m.Lock()
	for _, bridge := range bridges.Bridges {
		// For now, we only care about unallocated bridges.
		if bridge.Distributor != DistributorUnallocated {
			continue
		}

		if len(bridge.Transports) > 0 {
			for _, transport := range bridge.Transports {
				if transport.IsProbingResistant() {
					resp[transport.GetID()] = transport
					break
				}
			}
		} else {
			resp[bridge.GetID()] = bridge
		}

		i++
		if i == MaxReturnedBridges {
			break
		}
	}
	bridges.m.Unlock()

	return &resp, nil
}
