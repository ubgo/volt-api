package httpreq

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type GraphErr struct {
	Message string
}

func (e GraphErr) Error() string {
	return "graphql: " + e.Message
}

type GraphResponse struct {
	Data   interface{}
	Errors []GraphErr
	Error  struct {
		Message string `json:"message"`
	} `json:"error"`

	Extensions Extensions `json:"extensions"`
}

type ThrottleStatus struct {
	MaximumAvailable   float64 `json:"maximumAvailable"`
	CurrentlyAvailable int     `json:"currentlyAvailable"`
	RestoreRate        float64 `json:"restoreRate"`
}
type Cost struct {
	RequestedQueryCost int            `json:"requestedQueryCost"`
	ActualQueryCost    int            `json:"actualQueryCost"`
	ThrottleStatus     ThrottleStatus `json:"throttleStatus"`
}
type Extensions struct {
	Cost Cost `json:"cost"`
}

type GraphRequestBody struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

type GraphRequestParam struct {
	EndPoint string            `json:"endPoint"`
	Body     GraphRequestBody  `json:"body"`
	Headers  map[string]string `json:"headers"`
}

func RequestQuery(grp GraphRequestParam, resp interface{}) (*GraphResponse, error) {

	rp := RequestParam{
		EndPoint: grp.EndPoint,
		Headers:  grp.Headers,
		ArgsMap: map[string]interface{}{
			"query":     grp.Body.Query,
			"variables": grp.Body.Variables,
		},
	}

	graphResp := &GraphResponse{
		Data: &resp,
	}

	httpres, err := Do(rp, &graphResp)
	if err != nil {
		return nil, err
	}

	if len(graphResp.Errors) > 0 {
		return graphResp, graphResp.Errors[0]
	}

	if httpres.StatusCode != http.StatusOK {
		return graphResp, errors.New(graphResp.Error.Message)
	}

	return graphResp, nil
}

func GetItemTest() error {
	query := `
	query publicItem($search: String!) {
		publicItem(search: $search)
		{
				id
				title
		}
	}
	`
	rb := GraphRequestBody{
		Query: query,
		Variables: map[string]interface{}{
			"search": "test",
		},
	}

	grp := GraphRequestParam{
		EndPoint: "http://localhost:2101/query",
		Body:     rb,
	}

	var res interface{}
	_, err := RequestQuery(grp, &res)
	fmt.Println("Error", err)

	return nil
}
