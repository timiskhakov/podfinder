//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

package itunes

import "net/http"

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}
