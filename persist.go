
package persist

import (
	"path";
	"io";
	"os";
	"syscall";
	"log";
	"bytes";
	"strings";
	// "rand";
	// "strings";
	// "strconv";
	// "json";
)

// type Model struct {Id string; Data json.Json; Refs map[string]string; Keys []string; Indexes []string};

type Model struct {Id string; Data map[string]string};

type PersistService struct {Path string};

func NewPersistService(path string) *PersistService {
	p := new(PersistService);
	p.Path = path;
	return p;
}

// TODO: malicious ids
func (ps *PersistService) Get(id string) (*Model, bool) {
	log.Stderrf("PS:GET:%s", id);
	mpath := path.Join(ps.Path, id);
	log.Stderrf("PS:GET:path:%s", mpath);
	bs, e := io.ReadFile(mpath);
	// TODO: check error more specifically (not found, etc.)
	if e != nil {
		log.Stderrf("PS:GET:e:%s", e);
		return nil, false;
	}
	else {
		return parse_model(id, bs), true;
	}
	panic("unreachable");
}

func parse_model(id string, bs []byte) *Model {
	log.Stderrf("PS:parse_model:%s:%s", id, bs);
	data := parse_model_data(bs);
	m := new(Model);
	m.Id = id;
	m.Data = data;
	return m;
}

func parse_model_data(bs []byte) map[string]string {
	data := make(map[string]string);
	str := bytes.NewBuffer(bs).String();
	kvs := strings.Split(str,"&",0);
	for ikv := range kvs {
		if len(kvs[ikv]) > 0 {
			kv := strings.Split(kvs[ikv],"=",2);
			k := kv[0];
			if len(k) > 0 {
				var v = "";
				if len(kv) > 1 {
					v = kv[1];
				}
				data[k] = v;
			} // else: no field key error?
		}
	
	}
	return data;
}

func (ps *PersistService) Del(id string) os.Error {
	// TODO: malicious id path
	log.Stderrf(":DEL:id:%s", id);
	mpath := path.Join(ps.Path, id);
	e := os.Remove(mpath);
	return e;
}

func (ps *PersistService) New(id string, data map[string]string) *Model {
	// TODO: malicious id path
	log.Stderrf(":NEW:id:%s", id);
	for k,v := range data {
		log.Stderrf("NEW:k:%s:v:%s", k, v);
	}

	m := new(Model);
	m.Id = id;
	m.Data = data;

	bs := unparse_model_data(m.Data);

	mpath := path.Join(ps.Path, id);
	e := io.WriteFile(mpath, bs, syscall.S_IWUSR|syscall.S_IRUSR); // 00200|00400); // os.O_WRONLY|os.O_CREATE|os.O_TRUNC);
	if e != nil {
		// TODO: better error handling
		return nil;
	}
	else {
		return m;
	}
	panic("unreachable");
}

func unparse_model_data(data map[string]string) []byte {
	lines := make([]string, len(data));
	var i = 0;
	for k,v := range data {
		lines[i] = strings.Join([]string{k,v}, "=");
		i = i + 1;
	}
	return strings.Bytes(strings.Join(lines,"&"));
}

