# service

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)]
(http://godoc.org/github.com/rtwire/mock/service)

Package service implements a mock RTWire service with HTTP endpoints as described in RTWire's [documentation](https://rtwire.com/docs).

The service struct implements http.Handler allowing it to be used in unit/integrations tests with parts of your own code that integrate with RTWire's service. It can be used in combination with the Go client provided by RTWire located here [https://github.com/rtwire/go/client](https://github.com/rtwire/go/client). A small example is located below, however more examples are located in the Go client repository.

```go
func TestRTWire(t *testing.T) {
    s := httptest.NewServer(service.New())
    defer s.Close()
	
    cl := client.New(http.DefaultClient, s.URL+"/v1/mainnet", "user", "pass")
    acc, err := cl.CreateAccount()
	if err != nil {
        t.Fatal(err)
    }
    if acc.ID == 0 {
        t.Fatal("uninitialized account")
    }
}
```
