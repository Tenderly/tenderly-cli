package payloads

type NetworksResponse []*NetworkResponse

type NetworkResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	EthereumNetworkID string `json:"ethereum_network_id"`
}
