package main

import (
	"context"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// List of cookie names.
const (
	sessionTokenName = "sessionToken"
	sessionIDName    = "SID"
)

// List of XML API endpoints.
const (
	xmlGetter = "/xml/getter.xml"
	xmlSetter = "/xml/setter.xml"
)

// MetricsClient is a general purpose client, that gets metrics from
// a remote source.
type MetricsClient interface {
	Login(ctx context.Context) error
	Logout(ctx context.Context) error
	GetMetrics(ctx context.Context, fn string, out any) error
}

// ConnectBox is a client for ConnectBox HTTP API.
type ConnectBox struct {
	http     *http.Client
	addr     string
	token    string
	username string
	password string
}

// NewConnectBox creates new ConnectBox client.
func NewConnectBox(addr, username, password string) (*ConnectBox, error) {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}

	_, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %s", addr)
	}

	z := ConnectBox{
		addr:     strings.TrimSuffix(addr, "/"),
		username: username,
		password: hashPassword(password),
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("init cookie jar: %w", err)
	}
	z.http = &http.Client{
		Jar: jar,
		// Don't follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &z, nil
}

// Login gets auth token and session ID for further interactions
// with ConnectBox.
func (z *ConnectBox) Login(ctx context.Context) error {
	// Send a request just to set initial token
	_, err := z.get(ctx, "/common_page/login.html")
	if err != nil {
		return fmt.Errorf("get initial token: %w", err)
	}

	args := xmlArgs{
		{"Username", z.username},
		{"Password", z.password},
	}
	resp, err := z.xmlRequest(ctx, xmlSetter, FnLogin, args)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	if !strings.HasPrefix(resp, "success") {
		return fmt.Errorf("invalid response: %s", resp)
	}

	var sid string
	for _, item := range strings.Split(resp, ";") {
		kv := strings.Split(item, "=")
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "SID" {
			sid = kv[1]
			break
		}
	}
	if sid == "" {
		return fmt.Errorf("no SID in response from router: %s", resp)
	}
	z.setCookie(sessionIDName, sid)

	return nil
}

// Logout closes current session. This is important because ConnectBox
// is a single user device.
func (z *ConnectBox) Logout(ctx context.Context) error {
	_, err := z.xmlRequest(ctx, xmlSetter, FnLogout, xmlArgs{})
	return err
}

// GetMetrics reads metrics using XML RPC function with `fn` code, and
// unmarshals it to `out`.
func (z *ConnectBox) GetMetrics(ctx context.Context, fn string, out any) error {
	resp, err := z.xmlRequest(ctx, xmlGetter, fn, xmlArgs{})
	if err != nil {
		return fmt.Errorf("get response: %w", err)
	}
	if err := xml.Unmarshal([]byte(resp), out); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	return nil
}

func (z *ConnectBox) getCookie(name string) string {
	u, _ := url.Parse(z.addr)
	for _, cookie := range z.http.Jar.Cookies(u) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func (z *ConnectBox) setCookie(name, value string) {
	u, _ := url.Parse(z.addr)
	z.http.Jar.SetCookies(u, []*http.Cookie{{Name: name, Value: value}})
}

func (z *ConnectBox) xmlRequest(
	ctx context.Context,
	path string,
	fn string,
	args xmlArgs,
) (string, error) {
	// Token and function must be first arguments
	args = append(
		xmlArgs{{"token", z.token}, {"fun", fn}},
		args...,
	)
	return z.post(ctx, path, args.Encode())
}

func (z *ConnectBox) get(ctx context.Context, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, z.addr+path, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := z.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	// Token must be updated after each request
	z.token = z.getCookie(sessionTokenName)

	return string(body), nil
}

func (z *ConnectBox) post(ctx context.Context, path, data string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		z.addr+path,
		strings.NewReader(data),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := z.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	// Token must be updated after each request
	z.token = z.getCookie(sessionTokenName)

	return string(body), nil
}

func hashPassword(p string) string {
	h := sha256.New()
	h.Write([]byte(p))
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

// xmlArgs is a helper type for ConnectBox XML RPC, which requires ordered
// url-encoded requests. For example, `token` field must be always at the
// first place.
type xmlArgs [][2]string

// Encode returns url-encoded string with all keys and values.
func (args xmlArgs) Encode() (s string) {
	for _, arg := range args {
		if len(s) > 0 {
			s += "&"
		}
		s += url.QueryEscape(arg[0]) + "=" + url.QueryEscape(arg[1])
	}
	return s
}
