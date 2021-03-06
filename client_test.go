package knox

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMockClient(t *testing.T) {
	p := "primary"
	a := []string{"active1", "active2"}
	k := NewMock(p, a)
	p1 := k.GetPrimary()
	if p1 != p {
		t.Fatalf("Expected %s : Got %s for primary key", p, p1)
	}
	r := k.GetActive()
	if len(r) != len(a) {
		t.Fatalf("For active keys: length %d should equal length %d", len(r), len(a))
	}
	for i := range a {
		if r[i] != a[i] {
			t.Fatalf("%s should equal %s", r[i], a[i])
		}
	}
}

// buildServer returns a server. Call Close when finished.
func buildServer(code int, body []byte, a func(r *http.Request)) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a(r)
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
}

func buildGoodResponse(data interface{}) ([]byte, error) {
	resp := &Response{
		Status:    "ok",
		Code:      OKCode,
		Host:      "test",
		Timestamp: 1234567890,
		Message:   "",
		Data:      data,
	}
	return json.Marshal(resp)

}

func TestGetKey(t *testing.T) {
	expected := Key{
		ID:          "testkey",
		ACL:         ACL([]Access{}),
		VersionList: KeyVersionList{},
		VersionHash: "VersionHash",
	}
	resp, err := buildGoodResponse(expected)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	srv := buildServer(200, resp, func(r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("%s is not GET", r.Method)
		}
		if r.URL.Path != "/v0/keys/testkey/" {
			t.Fatalf("%s is not %s", r.URL.Path, "/v0/keys/testkey/")
		}
	})
	defer srv.Close()

	cli := MockClient(srv.Listener.Addr().String())

	k, err := cli.GetKey("testkey")
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	if k.ID != expected.ID {
		t.Fatalf("%s does not equal %s", k.ID, expected.ID)
	}
	if len(k.ACL) != len(expected.ACL) {
		t.Fatalf("%d does not equal %d", len(k.ACL), len(expected.ACL))
	}
	if len(k.VersionList) != len(expected.VersionList) {
		t.Fatalf("%d does not equal %d", len(k.VersionList), len(expected.VersionList))
	}
	if k.VersionHash != expected.VersionHash {
		t.Fatalf("%s does not equal %s", k.VersionHash, expected.VersionHash)
	}

}

func TestGetKeys(t *testing.T) {
	expected := []string{"a", "b", "c"}
	resp, err := buildGoodResponse(expected)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	srv := buildServer(200, resp, func(r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("%s is not GET", r.Method)
		}
		if r.URL.Path != "/v0/keys/" {
			t.Fatalf("%s is not %s", r.URL.Path, "/v0/keys/")
		}
		if r.URL.RawQuery != "y=x" {
			t.Fatalf("%s is not %s", r.URL.RawQuery, "y=x")
		}
	})
	defer srv.Close()

	cli := MockClient(srv.Listener.Addr().String())

	k, err := cli.GetKeys(map[string]string{"y": "x"})
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	if len(k) != 3 {
		t.Fatalf("%d is not 3", len(k))
	}
	if k[0] != "a" {
		t.Fatalf("%s is not %s", k[0], "a")
	}
	if k[1] != "b" {
		t.Fatalf("%s is not %s", k[0], "b")
	}
	if k[2] != "c" {
		t.Fatalf("%s is not %s", k[0], "c")
	}
}

func TestCreateKey(t *testing.T) {
	expected := uint64(123)
	resp, err := buildGoodResponse(expected)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	srv := buildServer(200, resp, func(r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("%s is not POST", r.Method)
		}
		if r.URL.Path != "/v0/keys/" {
			t.Fatalf("%s is not %s", r.URL.Path, "/v0/keys/")
		}
		r.ParseForm()
		if r.PostForm["data"][0] != "ZGF0YQ==" {
			t.Fatalf("%s is not expected: %s", r.PostForm["data"][0], "ZGF0YQ==")
		}
		if r.PostForm["id"][0] != "testkey" {
			t.Fatalf("%s is not expected: %s", r.PostForm["id"][0], "testkey")
		}
		if r.PostForm["acl"][0] == "" {
			t.Fatalf("%s is empty", r.PostForm["acl"][0])
		}
	})
	defer srv.Close()

	cli := MockClient(srv.Listener.Addr().String())

	acl := ACL([]Access{
		Access{
			Type:       User,
			AccessType: Read,
			ID:         "test",
		},
	})

	badACL := ACL([]Access{
		Access{
			Type:       233,
			AccessType: 80927,
			ID:         "test",
		},
	})
	_, err = cli.CreateKey("testkey", []byte("data"), badACL)
	if err == nil {
		t.Fatal("error is nil")
	}

	k, err := cli.CreateKey("testkey", []byte("data"), acl)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	if k != expected {
		t.Fatalf("%d is not %d", k, expected)
	}
}

func TestAddVersion(t *testing.T) {
	expected := uint64(123)
	resp, err := buildGoodResponse(expected)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	srv := buildServer(200, resp, func(r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("%s is not POST", r.Method)
		}
		if r.URL.Path != "/v0/keys/testkey/versions/" {
			t.Fatalf("%s is not %s", r.URL.Path, "/v0/keys/testkey/versions/")
		}
		r.ParseForm()
		if r.PostForm["data"][0] != "ZGF0YQ==" {
			t.Fatalf("%s is not expected: %s", r.PostForm["data"][0], "ZGF0YQ==")
		}
	})
	defer srv.Close()

	cli := MockClient(srv.Listener.Addr().String())

	k, err := cli.AddVersion("testkey", []byte("data"))
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	if k != expected {
		t.Fatalf("%d is not %d", k, expected)
	}
}

func TestDeleteKey(t *testing.T) {
	expected := ""
	resp, err := buildGoodResponse(expected)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	srv := buildServer(200, resp, func(r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("%s is not DELETE", r.Method)
		}
		if r.URL.Path != "/v0/keys/testkey/" {
			t.Fatalf("%s is not %s", r.URL.Path, "/v0/keys/testkey/")
		}
	})
	defer srv.Close()

	cli := MockClient(srv.Listener.Addr().String())

	err = cli.DeleteKey("testkey")
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
}

func TestPutVersion(t *testing.T) {
	resp, err := buildGoodResponse("")
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	srv := buildServer(200, resp, func(r *http.Request) {
		if r.Method != "PUT" {
			t.Fatalf("%s is not PUT", r.Method)
		}
		if r.URL.Path != "/v0/keys/testkey/versions/123/" {
			t.Fatalf("%s is not %s", r.URL.Path, "/v0/keys/testkey/versions/123/")
		}
		r.ParseForm()
		if r.PostForm["status"][0] != "\"Primary\"" {
			t.Fatalf("%s is not expected: %s", r.PostForm["status"][0], "\"Primary\"")
		}
	})
	defer srv.Close()

	cli := MockClient(srv.Listener.Addr().String())

	err = cli.UpdateVersion("testkey", "123", 2342)
	if err == nil {
		t.Fatal("error is nil")
	}

	err = cli.UpdateVersion("testkey", "123", Primary)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
}

func TestPutAccess(t *testing.T) {
	resp, err := buildGoodResponse("")
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
	srv := buildServer(200, resp, func(r *http.Request) {
		if r.Method != "PUT" {
			t.Fatalf("%s is not PUT", r.Method)
		}
		if r.URL.Path != "/v0/keys/testkey/access/" {
			t.Fatalf("%s is not %s", r.URL.Path, "/v0/keys/testkey/access/")
		}
		r.ParseForm()
		if r.PostForm["access"][0] == "" {
			t.Fatalf("%s is empty", r.PostForm["access"][0])
		}
	})
	defer srv.Close()

	cli := MockClient(srv.Listener.Addr().String())

	a := &Access{
		Type:       User,
		AccessType: Read,
		ID:         "test",
	}

	badA := &Access{
		Type:       233,
		AccessType: 80927,
		ID:         "test",
	}

	err = cli.PutAccess("testkey", badA)
	if err == nil {
		t.Fatal("error is nil")
	}

	err = cli.PutAccess("testkey", a)
	if err != nil {
		t.Fatalf("%s is not nil", err)
	}
}
