package lenta

import (
	"fmt"
	"net/http"
	"net/url"
)

func StoreCookies(cfg StoreConfig) []*http.Cookie {
	missionMode := fmt.Sprintf(
		`{"t":"pickup","ids":false,"ma":{"i":%d,"a":"%s","t":"%s","ri":1,"mt":"HM","s":false}}`,
		cfg.ID, cfg.Alias, cfg.Title,
	)

	return []*http.Cookie{
		{Name: "App_Cache_CitySlug", Value: cfg.City},
		{Name: "App_Cache_MissionAddressMode", Value: url.QueryEscape(missionMode)},
		{Name: "agree_with_cookie", Value: "true"},
		{Name: "Is_Search_Bot", Value: "false"},
	}
}

func StoreCookiesWithQrator(cfg StoreConfig, qratorJSID string) []*http.Cookie {
	cookies := StoreCookies(cfg)
	if qratorJSID != "" {
		cookies = append(cookies, &http.Cookie{
			Name:  "qrator_jsid",
			Value: qratorJSID,
		})
	}
	return cookies
}

type StoreConfig struct {
	ID    int
	Alias string
	Title string
	City  string
}
