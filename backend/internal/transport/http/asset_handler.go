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
	RespondOK(c, http.StatusOK, h.assetService.ListAssets())
}

func (h *AssetHandler) GetAsset(c *gin.Context) {
	asset, err := h.assetService.GetAsset(c.Param("assetId"))
	if err != nil {
		RespondError(c, http.StatusNotFound, "ASSET_NOT_FOUND", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, asset)
}
