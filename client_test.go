package gomailer

import "testing"

func TestNewClient(t *testing.T) {
	c, err := NewClient(Gomail, &Config{})
	if nil != err {
		t.Error("err should be nil, but got", err)
	}
	if nil == c {
		t.Error("client should not be nil")
	}

	c, err = NewClient(2, &Config{})
	if nil == err {
		t.Error("err should not be nil")
	}
	if nil != c {
		t.Error("client should be nil, but got", c)
	}
}
