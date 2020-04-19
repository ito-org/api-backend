package main

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ito-org/go-backend/tcn"
)

// GetRouter returns the Gin router.
func GetRouter(port string, dbConnection *DBConnection) *gin.Engine {
	h := &TCNReportHandler{}

	r := gin.Default()
	r.POST("/tcnreport", h.postTCNReport)
	r.GET("/tcnreport", h.getTCNReport)
	return r
}

// TCNReportHandler implements the handler functions for the API endpoints.
// It also holds the database connection that's used by the handler functions.
type TCNReportHandler struct {
	// todo: db connection
}

func (h *TCNReportHandler) postTCNReport(c *gin.Context) {
	body := c.Request.Body
	data, err := ioutil.ReadAll(body)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read request body")
		return
	}

	signedReport, err := tcn.GetSignedReport(data)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
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
		c.String(http.StatusBadRequest, "Failed to verify data")
	}
}

func (h *TCNReportHandler) getTCNReport(c *gin.Context) {
	// TODO
	c.String(http.StatusOK, "Here's your TCN")
}
