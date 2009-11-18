
package urllib

import (
	"fmt";
	"sort";
	"utf8";
	"strings";
	"bytes";
)

var always_safe = strings.Bytes("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-");

func urlquoter(c int, safe string) []byte {
	safe_bytes := strings.Bytes(safe);
	c_bytes := make([]byte, utf8.RuneLen(c));
	utf8.EncodeRune(c, c_bytes);
	if bytes.Index(safe_bytes, c_bytes) != -1 {
		return c_bytes;
	}
	else {
		if bytes.Index(always_safe, c_bytes) != -1 {
			return c_bytes;
		}
		else {
			return strings.Bytes(fmt.Sprintf("%%%02X", c));
		}
	}
	panic("unreachable");
}

func Urlquote_safe(s string, safe string) string {
	var qs = bytes.NewBufferString("");
	for _, c := range s {
		qs.Write(urlquoter(c,safe));
	}
	return qs.String();
}

func Urlquote(s string) string {
	return Urlquote_safe(s, "");
}

func SpaceToPlus(c int) int {
	if c == ' ' { return '+'; }
	else { return c; }
	panic("unreachable");
}

func Urlquote_plus_safe(s string, safe string) string {
	if strings.Index(s, " ") != -1 {
		s = Urlquote_safe(s, safe+" ");
		return strings.Map(SpaceToPlus,s);
	}
	else {
		return Urlquote_safe(s, safe);
	}
	panic("unreachable");
}

func Urlquote_plus(s string) string {
	return Urlquote_plus_safe(s, "");
}

func Urlencode(params map[string]string) string {
	keys := make([]string, len(params));
	i := 0;
	for k,_ := range params {
		keys[i] = k;
		i = i + 1;
	}
	sort.SortStrings(keys);

	sparamss := make([]string,len(params));
	j := 0;
	for k := range keys {
		sparamss[j] = Urlquote_plus(keys[k]) + "=" + Urlquote_plus(params[keys[k]]);
		j = j + 1;
	}
	sparams := strings.Join(sparamss, "&");
	return sparams;
}

