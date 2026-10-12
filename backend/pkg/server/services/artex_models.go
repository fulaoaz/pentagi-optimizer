package services

import (
	"net/http"
	"slices"

	"database/sql"
	"github.com/gin-gonic/gin"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/integrations/artex"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/server/response"
)

func NewArtexModelSharing(client *artex.Client, db database.Querier, sqlDB *sql.DB, cfg *config.Config, defaults provider.ProvidersConfig) *artex.ModelSharing {
	source := &artexModelSource{db: db, cfg: cfg, defaults: defaults}
	return artex.NewModelSharing(client, artex.NewSQLBindingStore(sqlDB), source, cfg.InstallationID+"/"+cfg.TenantID, cfg.AuthSalt())
}
func (s *ArtexService) ModelProviders(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	if !slices.Contains(c.GetStringSlice("prm"), "settings.providers.view") {
		response.Error(c, response.NewHttpError(403, "Artex.Forbidden", "Provider viewing permission required"), nil)
		return
	}
	items, err := s.sharing.ListProviders(c.Request.Context(), int64(c.GetUint64("uid")))
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	response.Success(c, http.StatusOK, items)
}
func (s *ArtexService) ModelBindings(c *gin.Context) {
	if !s.canView(c) {
		return
	}
	items, err := s.sharing.ListBindings(c.Request.Context(), int64(c.GetUint64("uid")))
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	response.Success(c, http.StatusOK, items)
}
func (s *ArtexService) SyncModel(c *gin.Context) {
	if !s.canWrite(c, "settings.providers.edit") {
		return
	}
	var input artex.ModelSelection
	if !bindArtexRequest(c, &input) {
		return
	}
	binding, err := s.sharing.Sync(c.Request.Context(), int64(c.GetUint64("uid")), input)
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	response.Success(c, http.StatusOK, binding)
}
func (s *ArtexService) RotateModelKey(c *gin.Context) {
	if !s.canWrite(c, "settings.providers.edit") {
		return
	}
	binding, err := s.sharing.RotateKey(c.Request.Context(), int64(c.GetUint64("uid")), c.Param("bindingID"))
	if err != nil {
		response.Error(c, artexError(err), nil)
		return
	}
	response.Success(c, http.StatusOK, binding)
}
