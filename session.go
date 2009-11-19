
package session

import (
	"http";
	"rand";
	"strings";
	"strconv";
	"log";
	"persist";
	// "json";
)

// type Session struct {Id string; Data map[string]string}; // Data was json.Json

// TODO: migrate in-memory session map to persistence
type SessionService struct {Name string; persist_service *persist.PersistService}; // sessions map[string]Session};

// e.g. name "MyService-Id"
func NewSessionService(name string, ps *persist.PersistService) *SessionService {
	ss := new(SessionService);
	ss.Name = name;
	ss.persist_service = ps;
	// ss.sessions = make(map[string]Session); // TODO: persistence
	return ss;
}

// TODO: synchronized for multithreading
// TODO: use real persistent sessions not in-memory map
// TODO: refactor out "cookie" handling
func (ss *SessionService) GetSession(c *http.Conn, req *http.Request) *persist.Model {
	log.Stderrf(">GetSession");
	cookie_val, has_cookie := req.Header["Cookie"];
	if has_cookie {
		log.Stderrf(":GetSession:has_cookie:%s", cookie_val);
		// TODO: this could crash server if malicious bad cookie is sent?
		// TODO: look for key Name-Id
		// sid=sid-0ce4603e-837d-4247-959e-6c17ef71d226; sid=sid-134020434
		var cookie_v = "";
		cookie_lines := strings.Split(cookie_val, "; ", 0);
		for cookie_i := range cookie_lines {
			cookie_line := cookie_lines[cookie_i];
			log.Stderrf("cookie_line:%s", cookie_line);
			cookie_kv := strings.Split(cookie_line, "=", 2);
			cookie_k := cookie_kv[0];
			if cookie_k == ss.Name {
				cookie_v = cookie_kv[1];
				break;
			}
		}
		if cookie_v == "" {
			log.Stderrf(":GetSession:no_cookie_v");
			return ss.StartSession(c, req, map[string]string{}); // json.Null);
		} else {
			s, has_session := ss.persist_service.Get(cookie_v);
				// ss.sessions[cookie_v];
			if has_session {
				log.Stderrf(":GetSession:has_session:%s", s.Id);
				return s;
			}
			else {
				log.Stderrf(":GetSession:no_session");
				return ss.StartSession(c, req, map[string]string{}); // json.Null);
			}
		}
	}
	else {
		log.Stderrf(":GetSession:no_cookie");
		return ss.StartSession(c, req, map[string]string{}); // json.Null);
	}
	panic("unreachable");
}

func (ss *SessionService) StartSession(c *http.Conn, req *http.Request, d map[string]string) *persist.Model { // d was json.Json
	log.Stderrf(">StartSession:");
	// TODO: uuid4 generate sid instead of "sid-" plus random
	// s := new(Session);
	// s.Id = ss.Name + "-" + strconv.Itoa(rand.Int());
	// s.Data = d;
	s := ss.persist_service.New(ss.Name+"-"+strconv.Itoa(rand.Int()), d);
	// TODO: refactor out cookie things
	// TODO: cookie domain
	c.SetHeader("Set-Cookie", ss.Name+"=" + s.Id + "; expires=Fri, 31-Dec-2011 23:59:59 GMT; path=/; domain=sol.caveman.org");
	// TODO: real thing not unprotected (threadwise) in-memory only "sessions"
	// ss.sessions[s.Id] = *s;
	return s;
}

