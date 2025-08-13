package rabbitmq

import (
	"testing"
)

func TestConstants(t *testing.T) {
	if ResultExchange != "oj.ex.results" {
		t.Errorf("ResultExchange = %s, want oj.ex.results", ResultExchange)
	}
	if ResultRoutingKey != "submission.result" {
		t.Errorf("ResultRoutingKey = %s, want submission.result", ResultRoutingKey)
	}
	if StatusExchange != "oj.ex.status" {
		t.Errorf("StatusExchange = %s, want oj.ex.status", StatusExchange)
	}
	if StatusRoutingKey != "submission.status" {
		t.Errorf("StatusRoutingKey = %s, want submission.status", StatusRoutingKey)
	}
}

func TestClientClose(t *testing.T) {
	client := &Client{ch: nil, conn: nil}
	client.Close()
}
