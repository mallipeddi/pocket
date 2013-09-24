package pocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// auth API URLs
	fetchRequestTokenUrl string = "https://getpocket.com/v3/oauth/request"
	fetchAccessTokenUrl  string = "https://getpocket.com/v3/oauth/authorize"
	authorizationUrl     string = "https://getpocket.com/auth/authorize?%s"

	// item API URLs
	retrieveUrl string = "https://getpocket.com/v3/get"
	addUrl      string = "https://getpocket.com/v3/add"
	modifyUrl   string = "https://getpocket.com/v3/send"
)

type Client struct {
	ConsumerToken string
	AccessToken   string
	Username      string
	c             *http.Client
}

type Error struct {
	StatusCode int
	ErrorCode  int
	ErrorMsg   string
}

type SortKind int

const (
	SortNewest SortKind = iota
	SortOldest SortKind = iota
	SortTitle  SortKind = iota
	SortSite   SortKind = iota
)

type ContentType int

const (
	TypeArticle ContentType = iota
	TypeVideo   ContentType = iota
	TypeImage   ContentType = iota
)

type ItemState int

const (
	StateUnread  ItemState = iota
	StateArchive ItemState = iota
	StateAll     ItemState = iota
)

type AddRequest struct {
	url     string
	title   string
	tags    []string
	tweetId string
}

func (req *AddRequest) SetUrl(url string) *AddRequest {
	req.url = url
	return req
}

func (req *AddRequest) SetTitle(title string) *AddRequest {
	req.title = title
	return req
}

func (req *AddRequest) AddTags(tags []string) *AddRequest {
	req.tags = append(req.tags, tags...)
	return req
}

func (req *AddRequest) SetTweetId(id string) *AddRequest {
	req.tweetId = id
	return req
}

type RetrieveRequest struct {
	params map[string]string
}

func NewRetrieveRequest() *RetrieveRequest {
	req := new(RetrieveRequest)
	req.params = make(map[string]string)
	return req
}

func (req *RetrieveRequest) Sort(kind SortKind) *RetrieveRequest {
	switch kind {
	case SortNewest:
		req.params["sort"] = "newest"
	case SortOldest:
		req.params["sort"] = "oldest"
	case SortTitle:
		req.params["sort"] = "title"
	case SortSite:
		req.params["sort"] = "site"
	}
	return req
}

func (req *RetrieveRequest) SimpleItemInfo() *RetrieveRequest {
	req.params["detailType"] = "simple"
	return req
}

func (req *RetrieveRequest) CompleteItemInfo() *RetrieveRequest {
	req.params["detailType"] = "complete"
	return req
}

func (req *RetrieveRequest) OnlyContentType(kind ContentType) *RetrieveRequest {
	switch kind {
	case TypeArticle:
		req.params["contentType"] = "article"
	case TypeVideo:
		req.params["contentType"] = "video"
	case TypeImage:
		req.params["contentType"] = "image"
	}
	return req
}

func (req *RetrieveRequest) OnlyTag(tag string) *RetrieveRequest {
	req.params["tag"] = tag
	return req
}

func (req *RetrieveRequest) OnlyUntagged() *RetrieveRequest {
	req.params["tag"] = "_untagged_"
	return req
}

func (req *RetrieveRequest) OnlyFavorited() *RetrieveRequest {
	req.params["favorite"] = "1"
	return req
}

func (req *RetrieveRequest) OnlyUnFavorited() *RetrieveRequest {
	req.params["favorite"] = "0"
	return req
}

func (req *RetrieveRequest) OnlyState(state ItemState) *RetrieveRequest {
	switch state {
	case StateUnread:
		req.params["state"] = "unread"
	case StateArchive:
		req.params["state"] = "archive"
	case StateAll:
		req.params["state"] = "all"
	}
	return req
}

func (req *RetrieveRequest) Count(count int) *RetrieveRequest {
	req.params["count"] = string(count)
	return req
}

func (req *RetrieveRequest) Offset(off int) *RetrieveRequest {
	req.params["offset"] = string(off)
	return req
}

func (req *RetrieveRequest) Since(timestamp string) *RetrieveRequest {
	req.params["since"] = timestamp
	return req
}

func (req *RetrieveRequest) OnlyDomain(domain string) *RetrieveRequest {
	req.params["domain"] = domain
	return req
}

func (req *RetrieveRequest) Search(key string) *RetrieveRequest {
	req.params["search"] = key
	return req
}

type ActionKind string

const (
	// basic actions
	ActionAdd        ActionKind = "add"
	ActionArchive    ActionKind = "archive"
	ActionReadd      ActionKind = "readd"
	ActionFavorite   ActionKind = "favorite"
	ActionUnfavorite ActionKind = "unfavorite"
	ActionDelete     ActionKind = "delete"

	// tagging actions
	ActionTagsAdd     ActionKind = "tags_add"
	ActionTagsRemove  ActionKind = "tags_remove"
	ActionTagsReplace ActionKind = "tags_replace"
	ActionTagsClear   ActionKind = "tags_clear"
	ActionTagRename   ActionKind = "tag_rename"
)

type Action struct {
	Kind   ActionKind
	Params map[string]string
}

type ModifyRequest struct {
	actions []Action
}

func (req *ModifyRequest) AddAction(a Action) {
	req.actions = append(req.actions, a)
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.ErrorCode, e.ErrorMsg)
}

func NewClient(consumerToken string) *Client {
	c := &http.Client{}
	client := &Client{ConsumerToken: consumerToken, c: c}
	return client
}

func NewClientWithAccessToken(consumerToken string, accessToken string, username string) *Client {
	c := &http.Client{}
	client := &Client{ConsumerToken: consumerToken, c: c, AccessToken: accessToken, Username: username}
	return client
}

func (client *Client) NewRequestToken(redirectUri string) (string, error) {
	var requestToken string

	v := url.Values{}
	v.Set("consumer_key", client.ConsumerToken)
	v.Set("redirect_uri", redirectUri)
	respStr, err := client.performPost(fetchRequestTokenUrl, v)
	if err != nil {
		return requestToken, err
	}

	respValues, err := url.ParseQuery(respStr)
	if err != nil {
		return requestToken, fmt.Errorf("Error parsing http response: %s", err)
	}
	requestToken = respValues.Get("code")
	return requestToken, nil
}

func (client *Client) GetAuthorizationUrl(requestToken string, redirectUri string) string {
	v := url.Values{}
	v.Set("request_token", requestToken)
	v.Set("redirect_uri", redirectUri)
	return fmt.Sprintf(authorizationUrl, v.Encode())
}

func (client *Client) FetchAccessToken(requestToken string) error {
	v := url.Values{}
	v.Set("consumer_key", client.ConsumerToken)
	v.Set("code", requestToken)

	respStr, err := client.performPost(fetchAccessTokenUrl, v)
	if err != nil {
		return err
	}

	respValues, err := url.ParseQuery(respStr)
	if err != nil {
		return fmt.Errorf("Error parsing http response: %s", err)
	}
	client.AccessToken = respValues.Get("access_token")
	client.Username = respValues.Get("username")
	return nil
}

func (client *Client) Retrieve(req *RetrieveRequest) (map[string]interface{}, error) {
	if err := client.verifyAccessToken(); err != nil {
		return nil, err
	}

	req.params["consumer_key"] = client.ConsumerToken
	req.params["access_token"] = client.AccessToken
	return client.performPostJson(retrieveUrl, req.params)
}

func (client *Client) Add(req *AddRequest) (map[string]interface{}, error) {
	if err := client.verifyAccessToken(); err != nil {
		return nil, err
	}

	params := make(map[string]string)
	params["consumer_key"] = client.ConsumerToken
	params["access_token"] = client.AccessToken
	params["url"] = req.url

	if len(req.title) > 0 {
		params["title"] = req.title
	}
	if len(req.tags) > 0 {
		params["tags"] = strings.Join(req.tags, ",")
	}
	if len(req.tweetId) > 0 {
		params["tweet_id"] = req.tweetId
	}

	return client.performPostJson(addUrl, params)
}

func (client *Client) Modify(req *ModifyRequest) (map[string]interface{}, error) {
	if err := client.verifyAccessToken(); err != nil {
		return nil, err
	}

	var l []interface{}
	for _, a := range req.actions {
		m := make(map[string]string)
		m["action"] = (string)(a.Kind)
		for k, v := range a.Params {
			m[k] = v
		}
		l = append(l, m)
	}
	actionsJson, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("consumer_key", client.ConsumerToken)
	params.Set("access_token", client.AccessToken)
	params.Set("actions", string(actionsJson[:]))

	encodedUrl := fmt.Sprintf("%s?%s", modifyUrl, params.Encode())

	resp, err := client.c.Get(encodedUrl)
	if err != nil {
		return nil, err
	}
	respBytes, err := client.handleResp(resp)
	if err != nil {
		return nil, err
	}

	var r interface{}
	if err := json.Unmarshal(respBytes, &r); err != nil {
		return nil, fmt.Errorf("Error parsing http response: %s", err)
	}

	m := r.(map[string]interface{})
	return m, nil
}

// private methods

func (client *Client) verifyAccessToken() error {
	if len(client.AccessToken) > 0 {
		return nil
	} else {
		return fmt.Errorf("missing access token")
	}
}

func (client *Client) performPost(requestUrl string, params url.Values) (string, error) {
	var respStr string
	resp, err := client.c.PostForm(requestUrl, params)
	if err != nil {
		return respStr, err
	} else {
		respBytes, err := client.handleResp(resp)
		respStr = string(respBytes[:])
		return respStr, err
	}
}

func (client *Client) performPostJson(
	requestUrl string, params map[string]string) (map[string]interface{}, error) {
	paramsEncoded, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	resp, err := client.c.Post(requestUrl, "application/json", bytes.NewReader(paramsEncoded))
	if err != nil {
		return nil, err
	} else {
		respBytes, err := client.handleResp(resp)
		if err != nil {
			return nil, err
		}

		var r interface{}
		if err := json.Unmarshal(respBytes, &r); err != nil {
			return nil, fmt.Errorf("Error parsing http response: %s", err)
		}

		m := r.(map[string]interface{})
		return m, nil
	}
}

func (client *Client) handleResp(resp *http.Response) ([]byte, error) {
	respBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return respBytes, fmt.Errorf("Error parsing http response body: %s", err)
	}

	if resp.StatusCode == 200 {
		return respBytes, nil
	} else {
		pErr := &Error{StatusCode: resp.StatusCode}
		if errCodeStr := resp.Header.Get("X-Error-Code"); len(errCodeStr) > 0 {
			pErr.ErrorCode, err = strconv.Atoi(errCodeStr)
		}
		pErr.ErrorMsg = resp.Header.Get("X-Error")

		return respBytes, pErr
	}
}
