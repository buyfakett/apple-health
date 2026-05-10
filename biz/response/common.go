package response

type CommonResp struct {
	Code Code   `json:"code"`
	Msg  string `json:"msg"`
}
