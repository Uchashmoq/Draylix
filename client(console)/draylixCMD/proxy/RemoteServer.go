package proxy

const (
	RECEIVE_LEN = 1024 * 32
)

type RemoteServer struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	Token   string `json:"token"`
	Address string `json:"address"`
	Iv      []byte `json:"initialVector"`
	Ikey    []byte `json:"initialKey"`
}
