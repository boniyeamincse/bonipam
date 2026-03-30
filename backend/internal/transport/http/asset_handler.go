package http

import (
	"boni-pam/internal/domain"
	"boni-pam/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	assetService *service.AssetService
}

func NewAssetHandler(assetService *service.AssetService) *AssetHandler {
	return &AssetHandler{assetService: assetService}
}

func (h *AssetHandler) RegisterRoutes(group *gin.RouterGroup) {
	assets := group.Group("/assets")
	assets.POST("", h.CreateAsset)
	assets.GET("", h.ListAssets)
	assets.GET("/:assetId", h.GetAsset)
	assets.PUT("/:assetId/tags", h.UpdateAssetTagging)
	assets.POST("/:assetId/test-connection", h.TestConnection)
}

func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var req domain.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid asset create request")
		return
	}

	asset, err := h.assetService.CreateAsset(req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ASSET_CREATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusCreated, asset)
}

func (h *AssetHandler) ListAssets(c *gin.Context) {
	environment := c.Query("environment")
	owner := c.Query("owner")
	criticality := c.Query("criticality")
	group := c.Query("group")

	RespondOK(c, http.StatusOK, h.assetService.ListAssets(environment, owner, criticality, group))
}

func (h *AssetHandler) GetAsset(c *gin.Context) {
	asset, err := h.assetService.GetAsset(c.Param("assetId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "ASSET_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, asset)
}

func (h *AssetHandler) UpdateAssetTagging(c *gin.Context) {
	var req domain.UpdateAssetTaggingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid asset tagging request")
		return
	}

	asset, err := h.assetService.UpdateAssetTagging(c.Param("assetId"), req)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ASSET_TAG_UPDATE_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, asset)
}

func (h *AssetHandler) TestConnection(c *gin.Context) {
	var req domain.TestAssetConnectionRequest
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			RespondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid test connection request")
			return
		}
	}

	result, err := h.assetService.TestConnection(c.Param("assetId"), req.TimeoutSeconds)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ASSET_CONNECTION_TEST_FAILED", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, result)
}
