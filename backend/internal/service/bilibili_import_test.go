package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestExtractPriceFromText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  float64
		ok    bool
	}{
		{name: "plus", input: "高价 900+ 小马糕", want: 900, ok: true},
		{name: "range", input: "出 800-900", want: 900, ok: true},
		{name: "keyword", input: "价格：999 元", want: 999, ok: true},
		{name: "currency", input: "￥123.4", want: 123, ok: true},
		{name: "invalid_over_max", input: "2026", want: 0, ok: false},
		{name: "missing", input: "no price here", want: 0, ok: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := extractPriceFromText(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v (price=%v)", tc.ok, ok, got)
			}
			if !ok {
				return
			}
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestExtractCodeFromText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "bracketed", input: "代码【A123】 价格900", want: "A123"},
		{name: "keyword", input: "兑换码: abcd12", want: "ABCD12"},
		{name: "digits_only_rejected", input: "代码：2026", want: ""},
		{name: "chinese_keyword_not_code", input: "代码：小马糕", want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := extractCodeFromText(tc.input)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestGetMixinKeyAllowsAnonymousNavResponseWithWbiImages(t *testing.T) {
	t.Parallel()

	importer := &BilibiliImporter{
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != bilibiliAPIBase+"/x/web-interface/nav" {
					t.Fatalf("unexpected request url: %s", req.URL.String())
				}

				body := `{"code":-101,"message":"账号未登录","data":{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/1234567890abcdef1234567890abcdef.png","sub_url":"https://i0.hdslb.com/bfs/wbi/fedcba0987654321fedcba0987654321.png"}}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	key, err := importer.getMixinKey(context.Background())
	if err != nil {
		t.Fatalf("expected anonymous nav response to provide mixin key, got error: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-char mixin key, got %q (len=%d)", key, len(key))
	}
	if importer.mixinKey != key {
		t.Fatalf("expected importer cache to store mixin key %q, got %q", key, importer.mixinKey)
	}
}

func TestGetMixinKeyRejectsNavResponseWithoutWbiImages(t *testing.T) {
	t.Parallel()

	importer := &BilibiliImporter{
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				body := `{"code":-101,"message":"账号未登录","data":{"wbi_img":{"img_url":"","sub_url":""}}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	_, err := importer.getMixinKey(context.Background())
	if err == nil {
		t.Fatal("expected missing wbi images to return an error")
	}
}
