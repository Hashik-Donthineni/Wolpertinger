package main

// GetBridges queries our backend to find and return bridges that we want
// tested by censorship measurement platforms like OONI.
func GetBridges(req *BridgeRequest) (*Bridges, error) {

	bs := NewBridges()

	bridges.m.Lock()
	for _, bridge := range bridges.Bridges {
		if bridge.Distributor == DistributorUnallocated {
			bs.Add(bridge)
			break
		}
	}
	bridges.m.Unlock()

	return bs, nil
}
