package steamauth

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type OpenID struct {
	root      string
	returnURL string
	data      url.Values
}

const (
	steamLogin       = "https://steamcommunity.com/openid/login"
	openIDMode       = "checkid_setup"
	openIDNS         = "http://specs.openid.net/auth/2.0"
	openIDIdentifier = "http://specs.openid.net/auth/2.0/identifier_select"
)

var (
	validationRegexp = regexp.MustCompile("^https://steamcommunity.com/openid/id/[0-9]{15,25}$")
	digitsRegexp     = regexp.MustCompile("\\D+")
)

// NewOpenID attaches a new OpenID request to the incoming request
func NewOpenID(r *http.Request) *OpenID {
	id := new(OpenID)

	proto := "http://"
	if r.TLS != nil {
		proto = "https://"
	}
	id.root = proto + r.Host

	uri := r.RequestURI
	if i := strings.Index(uri, "openid"); i != -1 {
		uri = uri[0 : i-1]
	}
	id.returnURL = id.root + uri

	switch r.Method {
	case "POST":
		id.data = r.Form
	case "GET":
		id.data = r.URL.Query()
	}

	return id
}

// AuthURL returns the current login url
func (id OpenID) AuthURL() string {
	u, _ := url.Parse(steamLogin)
	q := make(url.Values)
	q.Add("openid.claimed_id", openIDIdentifier)
	q.Add("openid.identity", openIDIdentifier)
	q.Add("openid.mode", openIDMode)
	q.Add("openid.ns", openIDNS)
	q.Add("openid.realm", id.root)
	q.Add("openid.return_to", id.returnURL)
	u.RawQuery = q.Encode()
	url := u.String()
	return url
}

// ValidateAndGetID authenticates the user and returns steamID
func (id *OpenID) ValidateAndGetID() (string, error) {
	if id.data.Get("openid.mode") != "id_res" {
		return "", errors.New(`mode must equal to "id_res"`)
	}

	if id.data.Get("openid.return_to") != id.returnURL {
		return "", errors.New(`return_to did not match the url of the request`)
	}

	params := make(url.Values)
	params.Set("openid.assoc_handle", id.data.Get("openid.assoc_handle"))
	params.Set("openid.signed", id.data.Get("openid.signed"))
	params.Set("openid.sig", id.data.Get("openid.sig"))
	params.Set("openid.ns", id.data.Get("openid.ns"))

	split := strings.Split(id.data.Get("openid.signed"), ",")
	for _, item := range split {
		params.Set("openid."+item, id.data.Get("openid."+item))
	}
	params.Set("openid.mode", "check_authentication")

	resp, err := http.PostForm(steamLogin, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := strings.Split(string(content), "\n")
	if response[0] != "ns:"+openIDNS {
		return "", errors.New("provider responded with the wrong namespace")
	}
	if strings.HasSuffix(response[1], "false") {
		return "", errors.New("unable to validate openID")
	}

	openIDURL := id.data.Get("openid.claimed_id")
	if !validationRegexp.MatchString(openIDURL) {
		return "", errors.New("provider responded with invalid id")
	}

	return digitsRegexp.ReplaceAllString(openIDURL, ""), nil
}

// Mode returns the current openID Mode
func (id OpenID) Mode() string {
	return id.data.Get("openid.mode")
}
