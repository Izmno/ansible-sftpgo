package sftpgo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	sftpgo "github.com/sftpgo/sdk"
	"golang.org/x/oauth2"
)

type ClientConfig struct {
	BaseURL  string `json:"base_url"`
	Username string `json:"admin_username"`
	Password string `json:"admin_password"`
}

type Client struct {
	Username string
	Password string

	baseUrl    string
	token      *oauth2.Token
	httpClient *http.Client

	mu sync.Mutex
}

type Token struct {
	AccessToken string    `json:"access_token"`
	Expiry      time.Time `json:"expires_at"`
}

type Error struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

func (e *Error) Error() string {
	if e.Err == "" {
		return e.Message
	}

	if e.Message == "" {
		return e.Err
	}

	return fmt.Sprintf("%s (%s)", e.Message, e.Err)
}

func NewClient(baseUrl, adminUsername, adminPassword string) (*Client, error) {
	if _, err := url.Parse(baseUrl); err != nil {
		return nil, err
	}

	client := &Client{
		Username: adminUsername,
		Password: adminPassword,
		baseUrl:  baseUrl,
	}

	httpClient := new(http.Client)
	httpClient.Transport = client

	client.httpClient = httpClient
	return client, nil
}

func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	t, err := c.Login(req.Context())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get authorization token: %v\n", err)

		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.AccessToken)

	return http.DefaultClient.Do(req)
}

func (c *Client) Login(ctx context.Context) (*oauth2.Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token.Valid() {
		return c.token, nil
	}

	t, err := c.GetNewToken(ctx)
	if err != nil {
		return t, err
	}

	c.token = t

	return t, nil
}

func (c *Client) GetNewToken(ctx context.Context) (*oauth2.Token, error) {
	res, err := resty.New().
		R().
		SetContext(ctx).
		SetBasicAuth(c.Username, c.Password).
		SetResult(&Token{}).
		SetError(&Error{}).
		Get(c.getUrl("api/v2/token"))

	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, res.Error().(*Error)
	}

	token := res.Result().(*Token)

	if token == nil || token.AccessToken == "" {
		return nil, errors.New("empty access token")
	}

	return &oauth2.Token{
		AccessToken: token.AccessToken,
		Expiry:      token.Expiry,
	}, nil
}

func (c *Client) GetUser(ctx context.Context, username string) (*sftpgo.BaseUser, error) {
	resp, err := resty.NewWithClient(c.httpClient).
		R().
		SetContext(ctx).
		SetBody(http.NoBody).
		SetResult(&sftpgo.BaseUser{}).
		SetError(&Error{}).
		Get(c.getUrl("api/v2/users", username))

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, nil
		}

		return nil, resp.Error().(*Error)
	}

	return resp.Result().(*sftpgo.BaseUser), nil
}

func (c *Client) CreateUser(ctx context.Context, user *sftpgo.BaseUser) error {
	resp, err := resty.NewWithClient(c.httpClient).
		R().
		SetContext(ctx).
		SetBody(user).
		SetResult(nil).
		SetError(&Error{}).
		Post(c.getUrl("api/v2/users"))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return resp.Error().(*Error)
	}

	return nil
}

func (c *Client) UpdateUser(ctx context.Context, user *sftpgo.BaseUser) error {
	resp, err := resty.NewWithClient(c.httpClient).
		R().
		SetContext(ctx).
		SetBody(user).
		SetResult(nil).
		SetError(&Error{}).
		Put(c.getUrl("api/v2/users", user.Username))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return resp.Error().(*Error)
	}

	return nil
}

func (c *Client) DeleteUser(ctx context.Context, username string) error {
	resp, err := resty.NewWithClient(c.httpClient).
		R().
		SetContext(ctx).
		SetBody(http.NoBody).
		SetResult(nil).
		SetError(&Error{}).
		Delete(c.getUrl("api/v2/users", username))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return resp.Error().(*Error)
	}

	return nil
}

func (c *Client) getUrl(pathSegments ...string) string {
	u, err := url.Parse(c.baseUrl)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(append([]string{u.Path}, pathSegments...)...)

	return u.String()
}
