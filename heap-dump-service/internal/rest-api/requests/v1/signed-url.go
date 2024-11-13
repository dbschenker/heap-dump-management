package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/config"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/metrics"
	"github.com/dbschenker/heap-dump-management/heap-dump-service/internal/rest-api/utils"
	"github.com/gin-gonic/gin"
)

type SigningRequest struct {
	Tenant    string `json:"tenant" example:"cloud-beacon"`
	Namespace string `json:"namespace" example:"beacon"`
	FileName  string `json:"filename" example:"test_file.dump"`
} // @name SigningRequest

type SigningResponse struct {
	URL                string `json:"url"`
	EncryptedAesKey    string `json:"encrypted-aes-key"`
	EncryptedAesKeyURL string `json:"encrypted-aes-key-url"`
	AesKey             string `json:"aes-key"`
} // @name SigningResponse

type ErrorResponse struct {
	Error string `json:"error"`
} // @name ErrorResponse

// @BasePath /api/v1

// @Summary Get signed upload URL
// @Schemes http https
// @Description Request a new Signed Upload URL for a specific file
// @Tags v1
// @param request body SigningRequest true "Request a new Signed Upload URL"
// @Accept json
// @Produce json
// @securityDefinitions.apikey ApiKeyAuth
// @Success      200  {object}  SigningResponse
// @Failure      400  {object}  ErrorResponse
// @Failure		 403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router /upload [post]
func HandleRequestUpload(c *gin.Context) {

	cfg := c.MustGet("cfg").(*config.AppConfig)
	vaultClient, err := utils.GenerateTransitVaultClient(cfg.Vault.VaultRole, cfg.Vault.VaultAuthMountPath, cfg.ServiceAccount.JWTokenMountPoint)

	if err != nil {
		log.WithFields(log.Fields{
			"caller": "HandleRequestUpload",
		}).Fatalf(fmt.Sprintf("unable to initialize Transit Vault Client : %s", err.Error()))
	}

	var requestBody SigningRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		errResp := ErrorResponse{
			Error: fmt.Sprintf("Could not Unmarshal request body %s", err.Error()),
		}
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	objectKey := fmt.Sprintf("%s/%s/%s", requestBody.Tenant, requestBody.Namespace, requestBody.FileName)
	aesKeyObjectKey := fmt.Sprintf("%s/%s/%s.%s", requestBody.Tenant, requestBody.Namespace, requestBody.FileName, "key")

	awsClient, err := utils.GenerateS3Client(cfg.App.Bucket)

	if err != nil {
		log.WithFields(log.Fields{
			"caller": "HandleRequestUpload",
		}).Error(fmt.Sprintf("Error initializing the AWS awsClient: %s", err.Error()))

		errResp := ErrorResponse{
			Error: fmt.Sprintf("Error initializing the AWS awsClient: %s", err.Error()),
		}
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}
	log.WithFields(log.Fields{
		"caller": "HandleRequestUpload",
	}).Info(fmt.Sprintf("Received request to presign PutObject for %s", objectKey))
	sdkReq, _ := awsClient.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(cfg.App.Bucket),
		Key:    aws.String(objectKey),
	})
	u, _, err := sdkReq.PresignRequest(15 * time.Minute)

	if err != nil {
		log.WithFields(log.Fields{
			"caller": "HandleRequestUpload",
		}).Error(fmt.Sprintf("Error Creating Signed URL: %s", err.Error()))

		errResp := ErrorResponse{
			Error: fmt.Sprintf("Error Creating Signed URL: %s", err.Error()),
		}
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	aesKey, err := utils.GenerateRandomBytes(32)

	if err != nil {
		log.WithFields(log.Fields{
			"caller": "HandleRequestUpload",
		}).Error(fmt.Sprintf("Error generating password: %s", err.Error()))
		errResp := ErrorResponse{
			Error: fmt.Sprintf("Error generating password: %s", err.Error()),
		}
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	encodedAesKey := utils.EncodeKey(aesKey)

	encryptedAesKey, err := utils.TransitEncryptString(vaultClient, cfg.Vault.VaultTransitMount, requestBody.Tenant, encodedAesKey)

	if err != nil {
		log.WithFields(log.Fields{
			"caller": "HandleRequestUpload",
		}).Error(fmt.Sprintf("Error encrypting password: %s", err.Error()))
		errResp := ErrorResponse{
			Error: fmt.Sprintf("Error encrypting password: %s", err.Error()),
		}
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	sdkReq, _ = awsClient.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(cfg.App.Bucket),
		Key:    aws.String(aesKeyObjectKey),
	})
	aesKeyURL, _, err := sdkReq.PresignRequest(15 * time.Minute)

	if err != nil {
		log.WithFields(log.Fields{
			"caller": "HandleRequestUpload",
		}).Error(fmt.Sprintf("Error generating presigned upload URL: %s", err.Error()))
		errResp := ErrorResponse{
			Error: fmt.Sprintf("Error generating presigned upload URL: %s", err.Error()),
		}
		c.JSON(http.StatusInternalServerError, errResp)
		return
	}

	resp := SigningResponse{
		URL:                u,
		EncryptedAesKey:    encryptedAesKey,
		EncryptedAesKeyURL: aesKeyURL,
		AesKey:             encodedAesKey,
	}

	c.JSON(http.StatusOK, resp)

	namespaceString := strings.ReplaceAll(requestBody.Namespace, "-", "_")

	metrics.HeapDumpHandled.WithLabelValues(namespaceString, requestBody.Tenant).Inc()

}
