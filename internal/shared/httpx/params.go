package httpx

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-core-go/internal/shared/authctx"
)

// ParamID membaca path param ID positif. Jika tidak valid, response 400
// sudah ditulis dan handler cukup return.
func ParamID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid "+name)
		return 0, false
	}
	return id, true
}

// CurrentUserID mengambil user yang sedang login. Jika tidak ada,
// response 401 sudah ditulis dan handler cukup return.
func CurrentUserID(c *gin.Context) (int64, bool) {
	id, ok := authctx.UserIDFrom(c.Request.Context())
	if !ok {
		Unauthorized(c)
	}
	return id, ok
}
