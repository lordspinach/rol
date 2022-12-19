package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"rol/app/errors"
	"strings"
	"time"
)

// 2016-09-27 09:38:21.541541811 +0200 CEST
// 127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700]
// "GET /apache_pb.gif HTTP/1.0" 200 2326
// "http://www.example.com/start.html"
// "Mozilla/4.08 [en] (Win98; I ;Nav)"

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (b bodyLogWriter) Write(bytes []byte) (int, error) {
	b.body.Write(bytes)
	return b.ResponseWriter.Write(bytes)
}

func notLog(path string, notLogged []string) bool {
	for _, n := range notLogged {
		if strings.Contains(path, n) {
			return true
		}
	}
	return false
}

func getOriginHeaders() map[string]bool {
	return map[string]bool{
		"Accept":                    true,
		"Accept-Encoding":           true,
		"Connection":                true,
		"Content-Length":            true,
		"Content-Type":              true,
		"User-Agent":                true,
		"Sec-Fetch-Dest":            true,
		"Accept-Language":           true,
		"Sec-Ch-Ua":                 true,
		"Sec-Ch-Ua-Platform":        true,
		"Sec-Ch-Ua-Mobile":          true,
		"Sec-Fetch-Site":            true,
		"Sec-Fetch-Mode":            true,
		"Sec-Fetch-User":            true,
		"Referer":                   true,
		"Cache-Control":             true,
		"Upgrade-Insecure-Requests": true,
	}
}

type handlerHelper struct {
	c      *gin.Context
	logger logrus.FieldLogger
	requestData
}

func newHandlerHelper(c *gin.Context, logger logrus.FieldLogger, stop time.Duration) *handlerHelper {
	h := &handlerHelper{
		c:      c,
		logger: logger,
	}
	h.requestData = h.getRequestData(stop)
	return h
}

type requestData struct {
	latency         int
	statusCode      int
	clientIP        string
	clientUserAgent string
	referer         string
	requestID       uuid.UUID
}

func (h *handlerHelper) getRequestData(stop time.Duration) requestData {
	return requestData{
		latency:         int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0)),
		statusCode:      h.c.Writer.Status(),
		clientIP:        h.c.ClientIP(),
		clientUserAgent: h.c.Request.UserAgent(),
		referer:         h.c.Request.Referer(),
	}
}

func (h *handlerHelper) getStringHeaders() (string, string) {
	originHeaders := getOriginHeaders()
	headers := h.c.Request.Header
	var headersString string
	var customHeadersString string
	for key, values := range headers {
		for _, value := range values {
			if originHeaders[key] {
				headersString = fmt.Sprint(headersString + key + ":" + value + " ")
			} else {
				customHeadersString = fmt.Sprint(customHeadersString + key + ":" + value + " ")
			}
		}
	}
	return headersString, customHeadersString
}

func (h *handlerHelper) getBytesAndRestoreBody() []byte {
	var bodyBytes []byte
	if h.c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(h.c.Request.Body)
	}
	// Restore the io.ReadCloser to its original state
	h.c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes
}

func (h *handlerHelper) getResponseHeaders() string {
	var respHeadersArr []string
	for header := range h.c.Writer.Header() {
		respHeadersArr = append(respHeadersArr, header)
	}
	var respHeaders string
	for _, value := range respHeadersArr {
		respHeaders += value + ":" + h.c.Writer.Header().Get(value) + " "
	}
	return respHeaders
}

func (h *handlerHelper) createEntry(stop time.Duration) *logrus.Entry {
	requestData := h.getRequestData(stop)
	bodyBytes := h.getBytesAndRestoreBody()
	respHeaders := h.getResponseHeaders()
	headersString, customHeadersString := h.getStringHeaders()
	queryParams := h.c.Request.URL.Query().Encode()
	domain := h.c.Request.Host
	return h.logger.WithFields(logrus.Fields{
		"domain":          domain,
		"statusCode":      requestData.statusCode,
		"latency":         requestData.latency, // time to process
		"clientIP":        requestData.clientIP,
		"method":          h.c.Request.Method,
		"referer":         requestData.referer,
		"userAgent":       requestData.clientUserAgent,
		"queryParams":     queryParams,
		"headers":         headersString,
		"requestBody":     string(bodyBytes),
		"customHeaders":   customHeadersString,
		"responseHeaders": respHeaders,
		"requestID":       h.requestID,
	})
}

func (h *handlerHelper) handleEntry(entry *logrus.Entry) {
	if len(h.c.Errors) > 0 {
		internalEntry := h.logger.WithFields(logrus.Fields{
			"actionID": h.requestID,
		})
		err := h.c.Errors.Last().Err
		if file := errors.GetCallerFile(err); file != "" {
			internalEntry = internalEntry.WithField("file", file)
		}
		if line := errors.GetCallerLine(err); line != -1 {
			internalEntry = internalEntry.WithField("line", line)
		}
		internalEntry.Error(err.Error())
	}
	if h.c.Writer.Status() >= http.StatusInternalServerError {
		entry.Error()
	} else if h.c.Writer.Status() >= http.StatusBadRequest {
		entry.Warn()
	} else {
		entry.Info()
	}
}

//Logger is the logrus logger handler
func Logger(logger logrus.FieldLogger, notLogged ...string) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return func(c *gin.Context) {
		requestID := uuid.New()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID.String())
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()
		stop := time.Since(start)
		if notLog(path, notLogged) {
			return
		}
		helper := newHandlerHelper(c, logger, stop)
		helper.requestID = requestID
		entry := helper.createEntry(stop)
		entry = entry.WithField("responseBody", blw.body.String())
		entry = entry.WithField("hostname", hostname)
		entry = entry.WithField("path", path)
		helper.handleEntry(entry)
	}
}
