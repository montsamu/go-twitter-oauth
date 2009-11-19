
package main

import (
	"oauth";
	"session";
	"flag";
	"http";
	"io";
	// "fmt";
	"log";
	"strings";
	"template";
	"runtime";
	"bytes";
	"json";
	"persist";
)

var oauth_persist_service = persist.NewPersistService("tokens");

// TODO: pull this twitter conf. stuff out to configuration not compilation
var twitter_callback_url = "http://my.url/callback/twitter";
var twitter_consumer_key = "myconsumerkey";
var twitter_consumer_secret = "myconsumersecret";
var twitter_client = oauth.NewTwitterClient(oauth_persist_service, twitter_consumer_key, twitter_consumer_secret, "authenticate");

var yahoo_callback_url = "http://my.url/callback/yahoo";
var yahoo_consumer_key = "myconsumerkey";
var yahoo_consumer_secret = "myconsumersecret";
var yahoo_client = oauth.NewYahooClient(oauth_persist_service, yahoo_consumer_key, yahoo_consumer_secret);

var google_callback_url = "http://my.url/callback/google";
var google_consumer_key = "myconsumerkey";
var google_consumer_secret = "myconsumersecret";
var google_client = oauth.NewGoogleClient(oauth_persist_service, google_consumer_key, google_consumer_secret);

var session_persist_service = persist.NewPersistService("sessions");
var session_service = session.NewSessionService(session_persist_service, "Example-Id");

var addr = flag.String("addr", ":8080", "http service address")
var maxprocs = flag.Int("maxprocs", 4, "max server processes")
var fmap = template.FormatterMap{
	"html": template.HTMLFormatter,
	"url+html": UrlHtmlFormatter,
}

func UrlHtmlFormatter(w io.Writer, v interface{}, fmt string) {
	template.HTMLEscape(w, strings.Bytes(http.URLEscape(v.(string))));
}

// TODO: figure out live reloading for Handlers? Is this easily possible?
func main() {
	flag.Parse();
	runtime.GOMAXPROCS(*maxprocs);
	// TODO: use a map datastructure of some type? nah. this works.
	// BUG?: "/" basically passes everything which doesn't match anything else :(
	http.Handle("/", http.HandlerFunc(DEFAULT));
	http.Handle("/twitter/replies", http.HandlerFunc(TWITTER_REPLIES));
	http.Handle("/login/twitter", http.HandlerFunc(LOGIN_TWITTER));
	http.Handle("/callback/twitter", http.HandlerFunc(CALLBACK_TWITTER));
	http.Handle("/login/yahoo", http.HandlerFunc(LOGIN_YAHOO));
	http.Handle("/callback/yahoo", http.HandlerFunc(CALLBACK_YAHOO));
	http.Handle("/login/google", http.HandlerFunc(LOGIN_GOOGLE));
	http.Handle("/callback/google", http.HandlerFunc(CALLBACK_GOOGLE));
	http.Handle("/logout", http.HandlerFunc(LOGOUT));
	http.Handle("/static/", http.FileServer("./static/", "/static/"));
	http.Handle("/favicon.ico", http.FileServer("./static/", "/"));
	err := http.ListenAndServe(*addr, nil);
	if err != nil {
		log.Exit("ListenAndServe:", err);
	}
}

// TODO: cache strategy
// TODO: learn best practice for reading template file into string and parsing
// TODO: check error from ReadFile
// TODO: synchronized cache for multithreading
func get_template(s string) *template.Template {
	var templateBytes, _ = io.ReadFile(s);
	var templateBuffer = bytes.NewBuffer(templateBytes);
	var templateStr = templateBuffer.String();
	return template.MustParse(templateStr, fmap);
}

// TODO: need uuid for session
// TODO: cookie domain site param?
// TODO: cookie exparation
func DEFAULT(c *http.Conn, req *http.Request) {
	s := session_service.GetSession(c,req);
	log.Stderrf("session data id:%s", s.Id);
	for k,v := range s.Data {
		log.Stderrf("session kv:%s:%s", k, v);
	}
	var defaultTemplate = get_template("default.html");
	defaultTemplate.Execute(s.Id, c);
}

func TWITTER_REPLIES(c *http.Conn, req *http.Request) {
	log.Stderrf(">REPLIES:");
	s := session_service.GetSession(c,req);
	for k,v := range s.Data {
		log.Stderrf("session kv:%s:%s", k, v);
	}
	auth_token, atx := s.Data["oauth_token"];
	if atx {
		log.Stderrf("TOKEN FOUND!");
		auth_token_secret := s.Data["oauth_token_secret"];
		r, finalUrl, err := twitter_client.MakeRequest("http://twitter.com/statuses/mentions.json", map[string]string{"oauth_token":auth_token}, auth_token_secret, false); //{"since_id":s.last_reply_id})
		if err != nil {
			log.Stderrf(":REPLIES:err:%s",err);
		}
		else {
			log.Stderrf(":REPLIES:r:%s:finalUrl:%s", r, finalUrl);
			b, _ := io.ReadAll(r.Body);
			print ("REPLIES!");
			str := bytes.NewBuffer(b).String();
       			println (str);
			j, ok, errtok := json.StringToJson(str);
			log.Stderr("REPLIES:j:%s:ok:%s:errtok:%s", j, ok, errtok);
			c.Write(strings.Bytes(j.String()));
		}
	
	}
	else {
		log.Stderrf("NO TOKEN FOUND!");
		http.Redirect(c, "/login/twitter?returnto=/twitter/replies", http.StatusFound); // should be 303 instead of 302?
	}
}

// accepts ?returnto=/replies etc.
func LOGIN_TWITTER(c *http.Conn, req *http.Request) {
	log.Stderr("TWITTER!");
	var url = twitter_callback_url;
	returnto := req.FormValue("returnto");
	if returnto != "" {
		url = url + "?returnto=" + returnto;
	}
	authorization_url := twitter_client.GetAuthorizationUrl(url);
	log.Stderrf("LOGIN:authorization_url:%s", authorization_url);
	http.Redirect(c, authorization_url, http.StatusFound); // should be 303 instead of 302?
}

// accepts ?returnto=/replies etc.
func LOGIN_GOOGLE(c *http.Conn, req *http.Request) {
	log.Stderr("GOOGLE!");
	var url = google_callback_url;
	returnto := req.FormValue("returnto");
	if returnto != "" {
		url = url + "?returnto=" + returnto;
	}
	authorization_url := google_client.GetAuthorizationUrl(url);
	log.Stderrf("LOGIN:authorization_url:%s", authorization_url);
	http.Redirect(c, authorization_url, http.StatusFound); // should be 303 instead of 302?
}

// accepts ?returnto=/replies etc.
func LOGIN_YAHOO(c *http.Conn, req *http.Request) {
	log.Stderr("YAHOO!");
	var url = yahoo_callback_url;
	returnto := req.FormValue("returnto");
	if returnto != "" {
		url = url + "?returnto=" + returnto;
	}
	authorization_url := yahoo_client.GetAuthorizationUrl(url);
	log.Stderrf("LOGIN:authorization_url:%s", authorization_url);
	http.Redirect(c, authorization_url, http.StatusFound); // should be 303 instead of 302?
}

func LOGOUT(c *http.Conn, req *http.Request) {
	session_service.EndSession(c, req);
	http.Redirect(c, "/", http.StatusFound); // allow returnto?
}

func CALLBACK_YAHOO(c *http.Conn, req *http.Request) {
	log.Stderr("CALLBACK YAHOO!");
	CALLBACK(c, req, twitter_client);
}

func CALLBACK_TWITTER(c *http.Conn, req *http.Request) {
	log.Stderr("CALLBACK TWITTER!");
	CALLBACK(c, req, twitter_client);
}

func CALLBACK_GOOGLE(c *http.Conn, req *http.Request) {
	log.Stderr("CALLBACK GOOGLE!");
	CALLBACK(c, req, google_client);
}

// receives ?returnto=/replies etc.
func CALLBACK(c *http.Conn, req *http.Request, auth_client *oauth.AuthClient) {
	log.Stderr("CALLBACK!");
	req.ParseForm();
	for k,v := range req.Header {
		log.Stderrf("header:%s:%s", k, v);
	}
	for k,vs := range req.Form {
		log.Stderrf("form:%s::", k);
		for i := range vs {
			log.Stderrf("::%s", vs[i]);
		}
	}

	var auth_token = req.FormValue("oauth_token");
	var auth_verifier = req.FormValue("oauth_verifier");
	log.Stderrf("CALLBACK:auth_token:%s:", auth_token);
	log.Stderrf("CALLBACK:auth_verifier:%s:", auth_verifier);

	user_info := auth_client.GetUserInfo(auth_token, auth_verifier);

	log.Stderrf("USER_INFO:");
	for k,v := range user_info {
		log.Stderrf("k:%s v:%s", k, v);
	}

	session_service.StartSession(c, req, user_info);
	var url = "/";
	returnto := req.FormValue("returnto");
	if returnto != "" {
		url = returnto;
	}
	http.Redirect(c, url, http.StatusFound); // should be 303 instead of 302?
}

