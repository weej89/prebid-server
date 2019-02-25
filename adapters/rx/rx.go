package rx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/pbs"
)

var _ adapters.Bidder = &RXAdapter{}

type RXAdapter struct {
	http *adapters.HTTPAdapter
	URI  string
}

type rxBidExt struct {
	AdspotID int `json:"adspot_id,omitempty"`
}

func (a *RXAdapter) Name() string {
	return "rx"
}

func (a *RXAdapter) SkipNoCookies() bool {
	return false
}

func NewRXAdapter(config *adapters.HTTPAdapterConfig, endpoint string) *RXAdapter {
	return NewRXBidder(adapters.NewHTTPAdapter(config).Client, endpoint)
}

func NewRXBidder(client *http.Client, endpoint string) *RXAdapter {
	a := &adapters.HTTPAdapter{Client: client}
	return &RXAdapter{
		http: a,
		URI:  endpoint,
	}
}

// Call function for Adapter is DEPRECATED
func (a *RXAdapter) Call(ctx context.Context, req *pbs.PBSRequest, bidder *pbs.PBSBidder) (pbs.PBSBidSlice, error) {
	return nil, nil
}

func (a *RXAdapter) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
	numRequests := len(request.Imp)
	var err error
	errs := make([]error, 0, len(request.Imp))
	requestData := make([]*adapters.RequestData, 0, numRequests)
	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	reqImpCopy := request.Imp
	for _, imp := range reqImpCopy {
		var bidExt adapters.ExtImpBidder
		if err = json.Unmarshal(imp.Ext, &bidExt); err != nil {
			errs = append(errs, &errortypes.BadInput{
				Message: err.Error(),
			})
			continue
		}
		var rxExt openrtb_ext.ExtImpRX
		if err = json.Unmarshal(bidExt.Bidder, &rxExt); err != nil {
			errs = append(errs, &errortypes.BadInput{
				Message: err.Error(),
			})
			continue
		}
		request.Imp = []openrtb.Imp{imp}
		reqJSON, err := json.Marshal(request)
		if err != nil {
			return nil, []error{err}
		}
		reqData := &adapters.RequestData{
			Method:  "POST",
			Uri:     a.URI,
			Body:    reqJSON,
			Headers: headers,
		}
		requestData = append(requestData, reqData)
	}
	return requestData, errs
}

func (a *RXAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}
	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode)}
	}
	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: err.Error(),
		}}
	}
	var bidReq openrtb.BidRequest
	if err := json.Unmarshal(externalRequest.Body, &bidReq); err != nil {
		return nil, []error{err}
	}
	bidResponse := adapters.NewBidderResponseWithBidsCapacity(len(bidResp.SeatBid))
	bidType := openrtb_ext.BidTypeBanner
	for _, sb := range bidResp.SeatBid {
		for i := 0; i < len(sb.Bid); i++ {
			bid := sb.Bid[i]
			if bid.Price != 0 {
				bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
					Bid:     &bid,
					BidType: bidType,
				})
			}
		}
	}
	return bidResponse, nil
}
