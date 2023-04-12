package function

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	sdk "github.com/openfaas/go-sdk"

	"github.com/sethvargo/go-password/password"
)

// ZoomMeetingResponse is the response from the Zoom API's
// meetings endpoint.
type ZoomMeetingResponse struct {
	UUID     string `json:"uuid"`
	ID       int64  `json:"id"`
	JoinURL  string `json:"join_url"`
	StartURL string `json:"start_url"`
	Password string `json:"password"`
	Topic    string `json:"topic"`
}

// ZoomTokenResponse is the response from Zoom's OAuth endpoint
// when requesting a token for server to server authentication.
type ZoomTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

// createMeeting creates a Zoom meeting based upon the following resources:
// https://zoom.github.io/api/#authentication
// https://marketplace.zoom.us/docs/api-reference/zoom-api/methods/#operation/meetingCreate
// https://zoom.github.io/api/#create-a-meeting
func createMeeting(topic string) (ZoomMeetingResponse, error) {
	var z ZoomMeetingResponse

	accountId, err := sdk.ReadSecret("zoom-account-id")
	if err != nil {
		return z, err
	}

	clientId, err := sdk.ReadSecret("zoom-client-id")
	if err != nil {
		return z, err
	}

	clientSecret, err := sdk.ReadSecret("zoom-client-secret")
	if err != nil {
		return z, err
	}

	token, err := requestZoomToken(accountId, clientId, clientSecret)
	if err != nil {
		return z, err
	}

	p, err := password.Generate(8, 4, 0, false, false)
	if err != nil {
		return z, err
	}

	zreq := ZoomMeetingRequest{
		Type:     2,
		Topic:    topic,
		Password: p,
		Duration: 60,
	}

	createMeetingJson, err := json.Marshal(zreq)
	if err != nil {
		return z, err
	}

	userId := "me"

	req, err := http.NewRequest(http.MethodPost, "https://api.zoom.us/v2/users/"+userId+"/meetings",
		bytes.NewBuffer([]byte(strings.TrimSpace(string(createMeetingJson)))))
	if err != nil {
		return z, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return z, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusCreated {
		return z, fmt.Errorf("createMeeting: %d - %s", res.StatusCode, string(body))
	}

	var zoomCall ZoomMeetingResponse
	if err := json.Unmarshal(body, &zoomCall); err != nil {
		return z, err
	}

	return zoomCall, err
}

func requestZoomToken(accountId, clientId, clientSecret string) (ZoomTokenResponse, error) {
	uri := "https://zoom.us/oauth/token"

	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return ZoomTokenResponse{}, err
	}

	q := url.Values{}
	q.Add("grant_type", "account_credentials")
	q.Add("account_id", accountId)

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", clientId, clientSecret)))))

	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ZoomTokenResponse{}, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusOK {
		return ZoomTokenResponse{}, fmt.Errorf("requestZoomToken: %s", string(body))
	}

	var token ZoomTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return ZoomTokenResponse{}, err
	}

	return token, nil
}

type ZoomMeetingRequest struct {
	Topic    string `json:"topic"`
	Type     int    `json:"type"`
	Password string `json:"password"`
	Duration int    `json:"duration"` // duration is set in minutes
}
