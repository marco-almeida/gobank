package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/marco-almeida/mybank/internal"
	"github.com/marco-almeida/mybank/internal/pkg"
	"github.com/marco-almeida/mybank/internal/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
	OverridePermissionKey   = "override_permission"
)

// Authentication creates a gin middleware for authorization
func Authentication(tokenMaker token.Maker, rolesWithPermission []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, internal.RenderErrorResponse(err.Error()))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, internal.RenderErrorResponse(err.Error()))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, internal.RenderErrorResponse(err.Error()))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, internal.RenderErrorResponse(err.Error()))
			return
		}

		if !hasAuthorization(payload.Role, rolesWithPermission) {
			err := errors.New("forbidden")
			ctx.AbortWithStatusJSON(http.StatusForbidden, internal.RenderErrorResponse(err.Error()))
			return
		}

		ctx.Set(OverridePermissionKey, canOverridePermission(payload.Role, rolesWithPermission))

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

func hasAuthorization(userRole string, rolesWithPermission []string) bool {
	if userRole == pkg.AdminRole {
		return true
	}

	return contains(rolesWithPermission, userRole)
}

func canOverridePermission(userRole string, rolesWithPermission []string) bool {
	if userRole == pkg.AdminRole {
		return true
	}

	// if bankers can override, allow
	if userRole == pkg.BankerRole {
		if contains(rolesWithPermission, pkg.BankerRole) {
			return true
		}
	}
	return false
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
