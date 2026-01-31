package natsutil

import (
	"encoding/json"
	"lace/nvalidator"
)

type ErrorResp struct {
	Msg  string `json:"msg,omitempty"`
	Code string `json:"code,omitempty"` // auth_login, token_error
}

type Resp struct {
	Data       interface{} `json:"data"`
	Error      *ErrorResp  `json:"error"`
	StatusCode uint        `json:"statusCode,omitempty"` // 400|401
}

func CreateRespWithErr(msg, code string, statusCode uint) Resp {
	return Resp{
		StatusCode: statusCode,
		Error: &ErrorResp{
			Msg:  msg,
			Code: code,
		},
	}
}

func CreateRespWithData(data interface{}) Resp {
	return Resp{
		Data:       data,
		StatusCode: 200,
	}
}

// Pare and validate the Nats request to predefiend Struct
//
//	r := PipelineRunReq{}
//	err := ParseAndValidate(req.Data(), &r)
//	if err != nil {
//		natsResp.SendError(natsresp.WithError(err))
//		return
//	}
func ParseAndValidate(data []byte, result interface{}) error {
	err := json.Unmarshal(data, result)
	if err != nil {
		return err
	}

	nvalid := nvalidator.New()

	err = nvalid.ValidateStruct(result)
	if err != nil {
		return err
	}
	return nil
}
