
package oauth

import (
	"urllib";
	"persist";

	"os";
	"sort";
	"io";
	"http";
	"log";
	"rand";
	// "bufio";
	"strings";
	"strconv";
	"bytes";
	"time";
	"crypto/hmac";
	"encoding/base64";
	// "container/vector";
)

// type AuthToken struct {Service string; Token string; Secret string; Created *time.Time};

// TODO: discover/write some persistence layer
// TODO: real persistence, not in-memory, thread-unsafe? nonsense
// var tokens = vector.New(100);

func create_auth_token(ps *persist.PersistService, token string, secret string) *persist.Model { // *AuthToken {
	t := ps.New(token, map[string]string{"secret":secret});
	// t := new(AuthToken);
	// t.Token = token;
	// t.Secret = secret;
	// t.Created = time.UTC();
	// tokens.Push(*t); // TODO: obvious this is terrible "persistence"
	return t;
}

// TODO: obvious things this is terrible
func get_auth_secret(ps *persist.PersistService, auth_token string) string {
	t, found := ps.Get(auth_token);
	if found {
		return t.Data["secret"];
	}
	else {
		return "BAD_TOKEN"; // TODO better handling
	}
	panic("unreachable");
	// for n := range tokens.Data() {
	// 	t := tokens.At(n);
	// 	token, _ := t.(AuthToken);
	// 	if token.Token == auth_token { return token.Secret };
	// }
	// return "BAD_TOKEN"; // TODO: better handling this is terrible
}

type AuthClient struct {
	persist_service *persist.PersistService;
	service_name string; // e.g. "twitter" "yahoo" etc.
	consumer_key string;
	consumer_secret string;
	request_url string; // where to request tokens
	access_url string; // where to exchange access tokens
	authorization_url string; // where to authorize
}

func (c *AuthClient) get_auth_token(callback_url string) *persist.Model { // *AuthToken {
	r, finalUrl, err := c.MakeRequest(c.request_url, map[string]string{"oauth_callback":callback_url}, "", false);
	if r != nil {
		log.Stderrf("get_auth_token:status:%s:finalUrl:%s", r.Status, finalUrl);
		for k, v := range r.Header {
			log.Stderrf("header:%s:%s", k, v);
		}
	}
	else {
		log.Stderrf("get_auth_token:finalUrl:%s:err:%s", finalUrl, err.String());
	}

	kvmap := parse_response(r); // check for 503
	// oauth_callback_confirmed=true assert?
	token := create_auth_token( c.persist_service, kvmap["oauth_token"], kvmap["oauth_token_secret"] );
	return token;
}

func parse_response(r *http.Response) map[string]string {
        b, _ := io.ReadAll(r.Body);
	print ("RESULT:");
	s := bytes.NewBuffer(b).String();
	println (s);
	vals := strings.Split(s, "&", 0);
	kvmap := make(map[string]string, len(vals));
	for i := range vals { // TODO: this crashes server for bad response right now
		kv := strings.Split(vals[i], "=", 2);
		kvmap[kv[0]] = kv[1]; // breaks server here on 503 not avail
	}
	// TODO: close r.Body ?
	return kvmap;
}

func NewAuthClient(ps *persist.PersistService, service_name string, consumer_key string, consumer_secret string, request_url string, access_url string, authorization_url string) *AuthClient {
	c := new(AuthClient);
	c.persist_service = ps;
	c.consumer_key = consumer_key;
	c.consumer_secret = consumer_secret;
	c.request_url = request_url;
	c.access_url = access_url;
	c.authorization_url = authorization_url;
	return c;
}

// authorization_type: authenticate | authorize
func NewTwitterClient(ps *persist.PersistService, consumer_key string, consumer_secret string, authorization_type string) *AuthClient {
	return NewAuthClient(ps, "twitter", consumer_key, consumer_secret, "http://twitter.com/oauth/request_token", "http://twitter.com/oauth/access_token", "http://twitter.com/oauth/"+authorization_type);
}

func (c *AuthClient) GetAuthorizationUrl(callback_url string) string {
	token := c.get_auth_token(callback_url);
	log.Stderrf("get_authorization_url:token:%s:secret:%s", token.Id, token.Data["secret"]);
	return c.authorization_url + "?oauth_token=" + token.Id + "&oauth_callback=" + urllib.Urlquote(callback_url);
}

// TODO: all
func (c *AuthClient) GetUserInfo(auth_token string, auth_verifier string) map[string]string {
	// get secret
	auth_secret := get_auth_secret(c.persist_service, auth_token);
	log.Stderrf("AUTH_SECRET:%s", auth_secret); // should client error if not found
	r, finalUrl, err := c.MakeRequest(c.access_url, map[string]string{"oauth_token":auth_token, "oauth_verifier":auth_verifier}, auth_secret, false);
	if r != nil {
		log.Stderrf("get_access_token:status:%s:finalUrl:%s", r.Status, finalUrl);
		for k, v := range r.Header {
			log.Stderrf("header:%s:%s", k, v);
		}
	}
	else {
		log.Stderrf("get_access_token:finalUrl:%s:err:%s", finalUrl, err.String());
	}
	kvmap := parse_response(r); // check for 503
	return kvmap;
}

func Digest(key string, m string) string {
	myhash := hmac.NewSHA1(strings.Bytes(key));
	myhash.Write(strings.Bytes(m));
	signature := bytes.TrimSpace(myhash.Sum());
	digest := make([]byte, base64.StdEncoding.EncodedLen(len(signature)));
	base64.StdEncoding.Encode(digest, signature);
	digest_str := strings.TrimSpace(bytes.NewBuffer(digest).String());
	return digest_str;
}

// build the message to sign/digest
func BuildMessage(url string, params map[string]string) string {
        i := 0;
	keys := make([]string,len(params));
        for k,_ := range params {
		keys[i] = k;
		i = i + 1;
	}
	sort.SortStrings(keys);

	j := 0;
        mss := make([]string,len(params));
	for k := range keys {
                mss[j] = urllib.Urlquote(keys[k]) + "=" + urllib.Urlquote(params[keys[k]]);
                j = j + 1;
        }
        ms := strings.Join(mss, "&");
        log.Stderrf("ms:%s", ms);

        m := strings.Join([]string{"GET", urllib.Urlquote(url), urllib.Urlquote(ms)}, "&");
        log.Stderrf("m:%s", m);

	return m;
}

func (c *AuthClient) MakeRequest(url string, additional_params map[string]string, token_secret string, protected bool) (r *http.Response, finalURL string, err os.Error) {

	log.Stderrf("make_request:url:%s:", url);
	for k,v := range additional_params {
		log.Stderrf("make_request:%s:%s:", k, v);
	}

	params := make(map[string]string);
	params["oauth_consumer_key"] = c.consumer_key;
	params["oauth_signature_method"] = "HMAC-SHA1";
	params["oauth_timestamp"] =  strconv.Itoa64(time.Seconds());
	params["oauth_nonce"] = strconv.Itoa64(rand.Int63());
	params["oauth_version"] = "1.0";

	//if token != "" {
	//	params["oauth_token"] = token;
	//}
	//else {
	//	params["oauth_callback"] = c.callback_url;
	//}

	// typically: oauth_token, oauth_callback, and/or oauth_verifier
	for k,v := range additional_params {
		params[k] = v;
	}

	for k,v := range params {
		log.Stderrf("param:%s:%s:", k, v);
	}

	m := BuildMessage(url, params);

	key := c.consumer_secret + "&" + token_secret;
	log.Stderrf("key:%s", key);

	digest_str := Digest(key, m);
	params["oauth_signature"] = digest_str;
	log.Stderrf("digest_str:%s", digest_str);

	sparams := urllib.Urlencode(params);
	log.Stderrf("sparams:%s", sparams);

	rurl := strings.Join([]string{url,sparams}, "?");

	log.Stderrf("make_request:rurl:%s", rurl);

	// TODO: no easy way to add header for Authorization:OAuth when protected...?
	return http.Get(rurl);
}

