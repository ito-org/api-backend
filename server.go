package main

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ito-org/go-backend/tcn"
)

const (
	requestBodyReadError    = "Failed to read request body"
	invalidRequestError     = "Invalid request"
	reportVerificationError = "Failed to verify report"
)

// GetRouter returns the Gin router.
func GetRouter(port string, dbConnection *DBConnection) *gin.Engine {
	h := &TCNReportHandler{
		dbConn: dbConnection,
	}

	r := gin.Default()
	r.POST("/tcnreport", h.postTCNReport)
	r.GET("/tcnreport", h.getTCNReport)
	return r
}

// TCNReportHandler implements the handler functions for the API endpoints.
// It also holds the database connection that's used by the handler functions.
type TCNReportHandler struct {
	dbConn *DBConnection
}

func (h *TCNReportHandler) postTCNReport(c *gin.Context) {
	body := c.Request.Body
	data, err := ioutil.ReadAll(body)
	if err != nil {
		c.String(http.StatusBadRequest, requestBodyReadError)
		return
	}

	signedReport, err := tcn.GetSignedReport(data)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if signedReport.Report.Memo == nil || signedReport.Report.Memo.Type != 0x2 {
		c.String(http.StatusBadRequest, invalidRequestError)
		return
	}

	ok, err := signedReport.Verify()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if ok {
		c.Status(http.StatusOK)
	} else {
		c.String(http.StatusBadRequest, reportVerificationError)
	}
}

func (h *TCNReportHandler) getTCNReport(c *gin.Context) {
	// TODO
	c.String(http.StatusOK, "Here's your TCN")
}
