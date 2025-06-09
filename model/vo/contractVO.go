package vo

type ContractVO struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Sequence int64  `json:"sequence"`
}
