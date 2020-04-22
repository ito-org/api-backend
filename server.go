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

	// If the memo field doesn't exist or the memo type is not ito's code, we
	// simply ignore the request.
	if signedReport.Report.Memo == nil || signedReport.Report.Memo.Type != 0x2 {
		c.String(http.StatusBadRequest, invalidRequestError)
		return
	}

	ok, err := signedReport.Verify()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if !ok {
		c.String(http.StatusBadRequest, reportVerificationError)
		return
	}

	if err := h.dbConn.insertSignedReport(signedReport); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (h *TCNReportHandler) getTCNReport(c *gin.Context) {
	reports, err := h.dbConn.getReports()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	data := []byte{}
	for _, r := range reports {
		b, err := r.Bytes()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		data = append(data, b...)
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}
