
package main

import (
	"oauth";
	"session";
	"persist";
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
)

var oauth_persist_service = persist.NewPersistService("tokens");

// TODO: pull this twitter conf. stuff out to configuration not compilation
var twitter_consumer_key = "my_consumer_key";
var twitter_consumer_secret = "my_consumer_secret";
var twitter_callback_url = "http://my.base.url/callback/twitter";
var twitter_client = oauth.NewTwitterClient(oauth_persist_service, twitter_consumer_key, twitter_consumer_secret, "authenticate");

var session_persist_service = persist.NewPersistService("sessions");
var session_service = session.NewSessionService(session_persist_service, "Example-Id");

var addr = flag.String("addr", ":8080", "http service address")
var maxprocs = flag.Int("maxprocs", 4, "max server processes")
var fmap = template.FormatterMap{
	"html": template.HTMLFormatter,
	"url+html": UrlHtmlFormatter,
}

// TODO: figure out live reloading for Handlers? Is this easily possible?
func main() {
	flag.Parse();
	runtime.GOMAXPROCS(*maxprocs);
	// TODO: use a map datastructure of some type? nah. this works.
	// BUG?: "/" basically passes everything which doesn't match anything else :(
	http.Handle("/", http.HandlerFunc(DEFAULT));
	http.Handle("/replies", http.HandlerFunc(REPLIES));
	http.Handle("/login/twitter", http.HandlerFunc(LOGIN_TWITTER));
	http.Handle("/callback/twitter", http.HandlerFunc(CALLBACK_TWITTER));
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
	log.Stderrf("session data:%s", s.Id);
	for k,v := range s.Data {
		log.Stderrf("session kv:%s:%s", k, v);
	}
	var defaultTemplate = get_template("default.html");
	defaultTemplate.Execute(s.Id, c);
}

func REPLIES(c *http.Conn, req *http.Request) {
	log.Stderrf(">REPLIES:");
	s := session_service.GetSession(c,req);
	auth_token, atx := s.Data["oauth_token"];
	if atx {
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
		http.Redirect(c, "/login/twitter?returnto=/replies", http.StatusFound); // should be 303 instead of 302?
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

func LOGOUT(c *http.Conn, req *http.Request) {
	session_service.EndSession(c, req);
	http.Redirect(c, "/", http.StatusFound); // allow returnto?
}

// twitter callback
// receives ?returnto=/replies etc.
func CALLBACK_TWITTER(c *http.Conn, req *http.Request) {
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

	user_info := twitter_client.GetUserInfo(auth_token, auth_verifier);

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

func UrlHtmlFormatter(w io.Writer, v interface{}, fmt string) {
	template.HTMLEscape(w, strings.Bytes(http.URLEscape(v.(string))));
}

