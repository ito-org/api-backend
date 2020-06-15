package main

import (
	"io/ioutil"
	"net/http"
	"math"
	"encoding/json"
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/openmined/tcn-psi/tcn"
	"github.com/openmined/tcn-psi/server"
)

const (
	requestBodyReadError    = "Failed to read request body"
	invalidRequestError     = "Invalid request"
	reportVerificationError = "Failed to verify report"
)

func GenerateSetupMessage(dbConnection *DBConnection, psicServer *server.TCNServer) (string, error) {
	var err error

	signedReports, err := dbConnection.getSignedReports()

	if err != nil {
		return "", err
	}

	fpr := math.Pow(10, -6)
	clientElements := int64(math.Pow(10, 4))

	var setupMessage string

	if len(signedReports) > 0 {
		setupMessage, err = psicServer.CreateSetupMessage(fpr, clientElements, signedReports)

		if err != nil {
			return "", err
		}
	} else {
		setupMessage = "{}"
	}

	return setupMessage, nil
}

// GetRouter returns the Gin router.
func GetRouter(port string, dbConnection *DBConnection, psicServer *server.TCNServer) *gin.Engine {
	setupMessage, err := GenerateSetupMessage(dbConnection, psicServer)

	if err != nil {
		return nil
	}

	h := &TCNReportHandler{
		setupMessage: setupMessage,
		psicServer: psicServer,
		dbConn: dbConnection,
	}

	r := gin.Default()
	r.POST("/publish", h.postTCNReport)
	r.GET("/setup", h.getSetupMessage)
	r.POST("/request", h.handleRequest)
	return r
}

// TCNReportHandler implements the handler functions for the API endpoints.
// It also holds the database connection that's used by the handler functions.
type TCNReportHandler struct {
	setupMessage string
	psicServer *server.TCNServer
	dbConn *DBConnection
}

func (h *TCNReportHandler) handleRequest(c *gin.Context) {
	var err error
	body := c.Request.Body
	data, err := ioutil.ReadAll(body)
	if err != nil {
		c.String(http.StatusBadRequest, requestBodyReadError)
		return
	}

	response, err := h.psicServer.ProcessRequest(string(data))

	if err != nil {
		c.String(http.StatusBadRequest, requestBodyReadError)
		return
	}

	c.Data(http.StatusOK, "application/json", []byte(response))
}

func (h *TCNReportHandler) postTCNReport(c *gin.Context) {
	body := c.Request.Body
	data, err := ioutil.ReadAll(body)
	if err != nil {
		c.String(http.StatusBadRequest, requestBodyReadError)
		return
	}

	var encodedTCNReports []string

	err = json.Unmarshal(data, &encodedTCNReports)

	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// TODO this will insert all reports until an error occurs. That might not be the desired behaviour.
	for _, encodedTCNReport := range encodedTCNReports {
		decodedReport, err := base64.StdEncoding.DecodeString(encodedTCNReport)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		signedReport, err := tcn.GetSignedReport(decodedReport)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		// If the memo type is not ito's code, we
		// simply ignore the request.
		if signedReport.Report.MemoType != 0x2 {
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
	}

	setupMessage, err := GenerateSetupMessage(h.dbConn, h.psicServer)

	if err == nil {
		h.setupMessage = setupMessage
	}

	c.Status(http.StatusOK)
}

func (h *TCNReportHandler) getSetupMessage(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", []byte(h.setupMessage))
}
