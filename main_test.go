package main

import (
	"strings"
	"testing"
)

func TestParseFailureReport(t *testing.T) {
	t.Parallel()

	logData := strings.NewReader(strings.Join([]string{
		`ok: [web-01] => {"changed":false,"msg":"all good"}`,
		`fatal: [db-01] | FAILED! => {"changed":false,"item":"pkg","msg":"package install failed"}`,
		`fatal: [db-02] | FAILED! => {"changed":true,"item":"svc","msg":"service restart failed"}`,
	}, "\n"))

	got, err := parseFailureReport(logData)
	if err != nil {
		t.Fatalf("parseFailureReport() error = %v", err)
	}

	want := "Host: db-01\nMessage: package install failed\n\nHost: db-02\nMessage: service restart failed\n\n"
	if got != want {
		t.Fatalf("parseFailureReport() = %q, want %q", got, want)
	}
}

func TestParseFailureReportIgnoresMalformedFailureLines(t *testing.T) {
	t.Parallel()

	logData := strings.NewReader(strings.Join([]string{
		`fatal: [db-01] | FAILED! => {not json}`,
		`fatal: [db-02] | FAILED!`,
		`fatal: [db-03] => {"changed":false,"msg":"missing host separator"}`,
	}, "\n"))

	got, err := parseFailureReport(logData)
	if err != nil {
		t.Fatalf("parseFailureReport() error = %v", err)
	}

	if got != "" {
		t.Fatalf("parseFailureReport() = %q, want empty string", got)
	}
}

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	got := string(buildMessage("ansible@example.com", "ops@example.com", "report body"))
	want := "From: ansible@example.com\nTo: ops@example.com\nSubject: Ansible Log Report\n\nreport body"
	if got != want {
		t.Fatalf("buildMessage() = %q, want %q", got, want)
	}
}

func TestExtractHostFallsBackToPipePrefix(t *testing.T) {
	t.Parallel()

	host, ok := extractHost([]byte("db-01 | failed: "))
	if !ok {
		t.Fatal("extractHost() did not find a host")
	}

	if got := string(host); got != "db-01" {
		t.Fatalf("extractHost() = %q, want %q", got, "db-01")
	}
}
