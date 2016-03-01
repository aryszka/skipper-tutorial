package main

import (
	"github.com/zalando/skipper"
	"github.com/zalando/skipper/filters"
	"log"
	"net/http"
)

type authSpec struct{}

type auth struct{}

func (as *authSpec) Name() string { return "auth" }

func (as *authSpec) CreateFilter(args []interface{}) (filters.Filter, error) {
	return &auth{}, nil
}

func unauthorized(ctx filters.FilterContext) {
	ctx.Serve(&http.Response{
		StatusCode: http.StatusUnauthorized,
		Header: http.Header{
			"WWW-Authenticate": []string{"Basic"}}})
}

func setToken(ctx filters.FilterContext) {
	ctx.Response().Header.Set("Set-Cookie", "auth=the-cake-is-a-lie;Path=/")
}

func (a *auth) Request(ctx filters.FilterContext) {
	if u, p, ok := ctx.Request().BasicAuth(); ok && u == "ken" && p == "thompson" {
		return
	}

	if c, err := ctx.Request().Cookie("auth"); err != nil || c.Value != "the-cake-is-a-lie" {
		unauthorized(ctx)
	}
}

func (a *auth) Response(ctx filters.FilterContext) {
	setToken(ctx)
}

func main() {
	log.Fatal(skipper.Run(skipper.Options{
		Address:       ":9090",
		CustomFilters: []filters.Spec{&authSpec{}},
		EtcdUrls:      []string{"http://localhost:2379"},
		EtcdPrefix:    "/skipper"}))
}
