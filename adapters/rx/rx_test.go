package rx

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/mxmCherry/openrtb"

	"github.com/prebid/prebid-server/adapters"
)

func TestRXAdapter_MakeRequests(t *testing.T) {
	bidder := new(RXAdapter)
	request := &openrtb.BidRequest{
		ID: "test-request-id",
		Imp: []openrtb.Imp{{
			ID: "test-imp-banner-id",
			Banner: &openrtb.Banner{
				Format: []openrtb.Format{{
					W: 728,
					H: 90,
				},
				},
			},
			Ext: json.RawMessage(`{"bidder": {
				"adspot_id": 1
			}}`),
		}},
		Device: &openrtb.Device{
			UA: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:65.0) Gecko/20100101 Firefox/65.0",
			Geo: &openrtb.Geo{
				Country: "US",
			},
			IP:       "127.0.0.1",
			Language: "en-US",
		},
		User: &openrtb.User{
			ID:       "test-user-id",
			BuyerUID: "test-buyer-id",
		},
	}
	reqs, errs := bidder.MakeRequests(request)
	if len(errs) > 0 {
		t.Errorf("got unexeptected errors while building HTTP requests: %v", errs)
	}
	if len(reqs) != 1 {
		t.Fatalf("unexepected number of HTTP requests.  Got %d. Expected %d.", len(reqs), 1)
	}
	for _, r := range reqs {
		if r.Method != "POST" {
			t.Errorf("Expected a POST message. Got %s", r.Method)
		}
		var rxReq openrtb.BidRequest
		if err := json.Unmarshal(r.Body, &rxReq); err != nil {
			t.Fatalf("Failed to unmarshal HTTP request: %v", err)
		}
		if rxReq.ID != request.ID {
			t.Errorf("Bad Request ID. Expected: %s, Got: %s", request.ID, rxReq.ID)
		}
		if len(rxReq.Imp) != len(request.Imp) {
			t.Fatalf("Wront len(request.Imp).  Expected: %d, Got %d", len(request.Imp), len(rxReq.Imp))
		}
		if rxReq.Cur != nil {
			t.Fatalf("Wrong request.Cur.  Epected: nil, Got: %s", rxReq.Cur)
		}
		if rxReq.Imp[0].ID == "test-imp-banner-id" {
			if rxReq.Imp[0].Banner.Format[0].W != 728 {
				t.Fatalf("Banner width does not match.  Expected: %d, Got: %d", 728, rxReq.Imp[0].Banner.Format[0].W)
			}
			if rxReq.Imp[0].Banner.Format[0].H != 90 {
				t.Fatalf("Banner height does not match.  Expected: %d, Got: %d", 90, rxReq.Imp[0].Banner.Format[0].H)
			}
		}
		if rxReq.Imp[0].Ext != nil {
			var bidderExt adapters.ExtImpBidder
			if err := json.Unmarshal(rxReq.Imp[0].Ext, &bidderExt); err != nil {
				t.Fatalf("Error unmarshalling bidder extension from request: %v", err)
			}
			var rxAdspotExt openrtb_ext.ExtImpRX
			if err := json.Unmarshal(bidderExt.Bidder, &rxAdspotExt); err != nil {
				t.Fatalf("Error unmarshalling adspot extension from bidder: %v", err)
			}
			if rxAdspotExt.AdspotId != 1 {
				t.Fatalf("Adspot ID does not match.  Expected: %d, Got: %d", 1, rxAdspotExt.AdspotId)
			}
		} else {
			t.Fatalf("BidExt object should not be nil and should contain a valid apspot_id.")
		}
	}
}

func TestOpenRTBEmptyResponse(t *testing.T) {
	httpResp := &adapters.ResponseData{
		StatusCode: http.StatusNoContent,
	}
	bidder := new(RXAdapter)
	bidReponse, errs := bidder.MakeBids(nil, nil, httpResp)
	if bidReponse != nil && len(bidReponse.Bids) != 0 {
		t.Errorf("Expected 0 bids. Got %d", len(bidReponse.Bids))
	}
	if len(errs) != 0 {
		t.Errorf("Expected 0 errors. Got %d", len(errs))
	}
}

func TestOpenRTBSurpriseReponse(t *testing.T) {
	httpResp := &adapters.ResponseData{
		StatusCode: http.StatusAccepted,
	}
	bidder := new(RXAdapter)
	bidReponse, errs := bidder.MakeBids(nil, nil, httpResp)
	if bidReponse != nil && len(bidReponse.Bids) != 0 {
		t.Errorf("Expected 0 bids. Got %d", len(bidReponse.Bids))
	}
	if len(errs) != 1 {
		t.Errorf("Expected 1 error. Got %d", len(errs))
	}
}

func TestOpenRTBStandardReponse(t *testing.T) {
	request := &openrtb.BidRequest{
		ID: "test-request-id",
		Imp: []openrtb.Imp{{
			ID: "test-imp-id",
			Banner: &openrtb.Banner{
				Format: []openrtb.Format{{
					W: 728,
					H: 90,
				}},
			},
			Ext: json.RawMessage(`{"bidder": {
				"adspotId": 1
			}}`),
		}},
	}
	requestJson, _ := json.Marshal(request)
	reqData := &adapters.RequestData{
		Method:  "POST",
		Uri:     "test-uri",
		Body:    requestJson,
		Headers: nil,
	}
	httpResp := &adapters.ResponseData{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"id":"test-request-id","seatbid":[{"bid":[{"id":"1234567890","impid":"test-imp-id","price":0.05664,"crid":"534348","adm":"test-ad","h":90,"w":728,"ext":{"bidder":{"seat":1}}}],"seat":"rbidder"}]}`),
	}
	bidder := new(RXAdapter)
	bidResponse, errs := bidder.MakeBids(request, reqData, httpResp)
	if bidResponse != nil && len(bidResponse.Bids) != 1 {
		t.Fatalf("Expected 1 bid. Got: %d", len(bidResponse.Bids))
	}
	if len(errs) != 0 {
		t.Errorf("Expected 0 errors. Got: %d", len(errs))
	}
	if bidResponse.Bids[0].BidType != openrtb_ext.BidTypeBanner {
		t.Errorf("Expected a banner bid. Got: %s", bidResponse.Bids[0].BidType)
	}
	theBid := bidResponse.Bids[0].Bid
	if theBid.ID != "1234567890" {
		t.Errorf("Bad bid ID. Expected: %s, Got: %s", "1234567890", theBid.ID)
	}
}
