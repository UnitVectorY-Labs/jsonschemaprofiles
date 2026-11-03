package main

import (
	"fmt"
	"runtime"
	"testing"
)

func TestBuildVersionOutputAddsVPrefixForSemver(t *testing.T) {
	got := buildVersionOutput("1.2.3")
	want := fmt.Sprintf("v1.2.3 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestBuildVersionOutputKeepsVPrefix(t *testing.T) {
	got := buildVersionOutput("v1.2.3")
	want := fmt.Sprintf("v1.2.3 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestBuildVersionOutputKeepsNonSemver(t *testing.T) {
	got := buildVersionOutput("dev")
	want := fmt.Sprintf("dev (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestBuildVersionOutputKeepsEmptyString(t *testing.T) {
	got := buildVersionOutput("")
	want := fmt.Sprintf(" (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestBuildVersionOutputKeepsPartialSemverWithoutPrefix(t *testing.T) {
	got := buildVersionOutput("1.2")
	want := fmt.Sprintf("1.2 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestBuildVersionOutputAddsVPrefixForFourPartVersion(t *testing.T) {
	got := buildVersionOutput("1.2.3.4")
	want := fmt.Sprintf("v1.2.3.4 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestBuildVersionOutputAddsVPrefixForPrereleaseSemver(t *testing.T) {
	got := buildVersionOutput("1.2.3-alpha+build")
	want := fmt.Sprintf("v1.2.3-alpha+build (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}
