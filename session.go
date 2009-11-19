
package session

import (
	"http";
	"rand";
	"strings";
	"strconv";
	// "json";
)

type Session struct {Sid string; Data map[string]string} // Data was json.Json

// TODO: persist, not simply a map in memory
// map of sid string to Session
var sessions = make(map[string]Session);

// TODO: synchronized for multithreading
// TODO: use real persistent sessions not in-memory map
// TODO: refactor out "cookie" handling
func GetSession(c *http.Conn, req *http.Request) *Session {
	cookie_val, has_cookie := req.Header["Cookie"];
	if has_cookie {
		// TODO: this could crash server if malicious bad cookie is sent?
		sid_val := strings.Split(cookie_val, "=", 2)[1];
		s, has_session := sessions[sid_val];
		if has_session {
			return &s;
		}
		else {
			return StartSession(c, req, map[string]string{}); // json.Null);
		}
	}
	else {
		return StartSession(c, req, map[string]string{}); // json.Null);
	}
	panic("unreachable");
}

func StartSession(c *http.Conn, req *http.Request, d map[string]string) *Session { // d was json.Json
	s := new(Session);
	// TODO: uuid4 generate sid instead of "sid-" plus random
	s.Sid = "sid-" + strconv.Itoa(rand.Int());
	s.Data = d;
	// TODO: refactor out cookie things
	c.SetHeader("Set-Cookie", "sid=" + s.Sid + "; expires=Fri, 31-Dec-2011 23:59:59 GMT; path=/; domain=sol.caveman.org");
	// TODO: real thing not unprotected (threadwise) in-memory only "sessions"
	sessions[s.Sid] = *s;
	return s;
}

