package controller

import (
	"net/http"
	"strconv"

	"auth-service/internal/entity"
	"auth-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHTTPController struct {
	authUC usecase.IAuthUseCase
}

func NewAuthHTTPController(authUC usecase.IAuthUseCase) *AuthHTTPController {
	return &AuthHTTPController{authUC: authUC}
}

type SignUpRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=user admin"`
}

type SignInRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignUp регистрирует нового пользователя
// @Summary Регистрация пользователя
// @Description Создает новую учетную запись пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "Данные для регистрации"
// @Success 201 {object} map[string]interface{} "id"
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /api/v1/auth/signup [post]
func (c *AuthHTTPController) SignUp(ctx *gin.Context) {
	var req SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := entity.UserRegister{
		Username: req.Username,
		Password: req.Password,
	}

	createdUser, err := c.authUC.Register(ctx.Request.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id": createdUser.ID,
	})
}

// SignIn аутентифицирует пользователя
// @Summary Вход в систему
// @Description Аутентификация пользователя по email и паролю
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignInRequest true "Данные для входа"
// @Success 200 {object} map[string]interface{} "access_token"
// @Failure 400 {object} entity.ErrorResponse
// @Failure 401 {object} entity.ErrorResponse
// @Router /api/v1/auth/signin [post]
func (c *AuthHTTPController) SignIn(ctx *gin.Context) {
	var req SignInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginReq := entity.UserLogin{
		Username: req.Username,
		Password: req.Password,
	}

	session, err := c.authUC.Login(ctx.Request.Context(), loginReq)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
	})
}

// GetUserProfile возвращает профиль пользователя
// @Summary Получить профиль
// @Description Возвращает информацию о пользователе
// @Tags Auth
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} entity.User
// @Failure 404 {object} entity.ErrorResponse
// @Router /api/v1/auth/users/{id} [get]
func (c *AuthHTTPController) GetUserProfile(ctx *gin.Context) {
	idStr := ctx.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := c.authUC.GetUserByID(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// ValidateToken проверяет валидность токена
// @Summary Проверить токен
// @Description Валидирует JWT токен
// @Tags Auth
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "user_id"
// @Failure 401 {object} entity.ErrorResponse
// @Router /api/v1/auth/validate [get]
func (c *AuthHTTPController) ValidateToken(ctx *gin.Context) {
	tokenStr := ctx.GetHeader("Authorization")
	if tokenStr == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
		return
	}

	parsedToken, err := c.authUC.ValidateToken(tokenStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id": claims["user_id"],
		"role":    claims["role"],
	})
}