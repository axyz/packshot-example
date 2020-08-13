package filters

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/axyz/packshot-example/tools"
	log "github.com/sirupsen/logrus"
	"github.com/zalando/skipper/filters"
)

const (
	PackshotName                = "packshot"
	InitialBufferFallbackLength = 500000
	InitialBufferExtraBytes     = 4096
)

type packshot struct {
}

func NewPackshot() filters.Spec {
	return &packshot{}
}

func (p *packshot) Name() string {
	return PackshotName
}

func (p *packshot) CreateFilter(args []interface{}) (filters.Filter, error) {
	f := &packshot{}

	return f, nil
}

func (p *packshot) Request(ctx filters.FilterContext) {
}

func (p *packshot) Response(ctx filters.FilterContext) {
	rsp := ctx.Response()

	// handle not 200 responses from loopback (e.g. cache or 404)
	if rsp.StatusCode != http.StatusOK {
		return
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Error("[packshot] failed to get source image ", err)
		rsp.StatusCode = 500
		return
	}
	rsp.Body.Close()

	output, err := tools.CreatePackshot(body)
	if err != nil {
		log.Error("[packhsot] Error while generating the packshot with ", err)
		rsp.StatusCode = 500
		return
	}

	rsp.Header.Del("Content-Length")

	rsp.Body = ioutil.NopCloser(bytes.NewReader(output))
}
