package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StartServer runs the HTTP server
func StartServer(port string, dbConnection *DBConnection) error {
	h := &TCNReportHandler{}

	r := gin.Default()
	r.POST("/tcnreport", h.postTCNReport)
	r.GET("/tcnreport", h.getTCNReport)
	return r.Run(fmt.Sprintf(":%s", port))
}

// TCNReportHandler implements the handler functions for the API endpoints.
// It also holds the database connection that's used by the handler functions.
type TCNReportHandler struct {
	// todo: db connection
}

func (h *TCNReportHandler) postTCNReport(c *gin.Context) {
	// TODO
	c.String(http.StatusOK, "POST that TCN")
}

func (h *TCNReportHandler) getTCNReport(c *gin.Context) {
	// TODO
	c.String(http.StatusOK, "Here's your TCN")
}
