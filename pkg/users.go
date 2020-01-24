package pkg

import (
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/staple-org/staple/internal/models"
	"github.com/staple-org/staple/internal/service"
	"github.com/staple-org/staple/pkg/config"
)

// RegisterUser takes a storer and creates a user entry.
func RegisterUser(userHandler service.UserHandlerer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the nickname for the token.
		user := &models.User{}
		err := c.Bind(user)
		if user.Email == "" || user.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid username or password",
			})
		}
		if err != nil {
			return err
		}
		if ok, err := userHandler.IsRegistered(*user); ok {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "User already registered.",
			})
		} else if err != nil {
			log.Println("[ERROR] During registration flow: ", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return userHandler.Register(*user)
	}
}

// ResetPassword takes a user handler and resets a user's password delievered from the token.
func ResetPassword(userHandler service.UserHandlerer) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := &models.User{}
		err := c.Bind(user)
		if err != nil {
			log.Println("[ERROR] Failed binding user: ", err)
			return err
		}
		if user.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "invalid email",
			})
		}
		if ok, err := userHandler.IsRegistered(*user); !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "User not found",
			})
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		return userHandler.SendConfirmCode(*user)
	}
}

// ChangePassword let's the user change the account's password.
func ChangePassword(userHandler service.UserHandlerer) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := GetToken(c)
		if err != nil {
			return err
		}
		claims := token.Claims.(jwt.MapClaims)
		email := claims["email"].(string)
		userModel := &models.User{
			Email: email,
		}

		var password = struct {
			Password string `json:"password"`
		}{}
		err = c.Bind(&password)
		if err != nil {
			return err
		}
		if password.Password == "" {
			apiError := config.ApiError("password is empty", http.StatusBadRequest, nil)
			return c.JSON(http.StatusBadRequest, apiError)
		}
		err = userHandler.ChangePassword(*userModel, password.Password)
		if err != nil {
			apiError := config.ApiError("failed to change password", http.StatusInternalServerError, nil)
			return c.JSON(http.StatusInternalServerError, apiError)
		}
		return c.NoContent(http.StatusOK)
	}
}

// SetMaximumStaples let's the user change the maximum number of staples.
func SetMaximumStaples(userHandler service.UserHandlerer) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := GetToken(c)
		if err != nil {
			return err
		}
		claims := token.Claims.(jwt.MapClaims)
		email := claims["email"].(string)
		userModel := &models.User{
			Email: email,
		}

		var maxStaples = struct {
			Staples string `json:"max_staples"`
		}{}
		err = c.Bind(&maxStaples)
		if err != nil {
			return err
		}
		stapleCount, err := strconv.Atoi(maxStaples.Staples)
		if err != nil {
			apiError := config.ApiError("failed to convert staple to string", http.StatusBadRequest, err)
			return c.JSON(http.StatusBadRequest, apiError)
		}
		if stapleCount <= 0 || stapleCount > 100 {
			apiError := config.ApiError("invalid staple setting", http.StatusBadRequest, nil)
			return c.JSON(http.StatusBadRequest, apiError)
		}
		err = userHandler.SetMaximumStaples(*userModel, stapleCount)
		if err != nil {
			apiError := config.ApiError("failed to set maximum staples", http.StatusInternalServerError, nil)
			return c.JSON(http.StatusInternalServerError, apiError)
		}
		return c.NoContent(http.StatusOK)
	}
}

// GetMaximumStaples returns the user's maximum staple count.
func GetMaximumStaples(userHandler service.UserHandlerer) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := GetToken(c)
		if err != nil {
			return err
		}
		claims := token.Claims.(jwt.MapClaims)
		email := claims["email"].(string)
		userModel := &models.User{
			Email: email,
		}
		staples, err := userHandler.GetMaximumStaples(*userModel)
		if err != nil {
			apiError := config.ApiError("failed to get maximum staples", http.StatusInternalServerError, nil)
			return c.JSON(http.StatusInternalServerError, apiError)
		}

		var maxStaples = struct {
			Staples string `json:"max_staples"`
		}{
			Staples: strconv.Itoa(staples),
		}

		return c.JSON(http.StatusOK, maxStaples)
	}
}

// VerfiyConfirmCode verifies a confirm link generated by a reset password action.
func VerfiyConfirmCode(userHandler service.UserHandlerer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the nickname for the token.
		var confirm = struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}{}
		err := c.Bind(&confirm)
		if err != nil {
			return err
		}
		if confirm.Email == "" || confirm.Code == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "invalid email or code",
			})
		}

		user := models.User{Email: confirm.Email, ConfirmCode: confirm.Code}
		if ok, err := userHandler.IsRegistered(user); !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "User not found",
			})
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		if ok, err := userHandler.VerifyConfirmCode(user); err != nil {
			apiError := config.ApiError("error while confirming link", http.StatusBadRequest, err)
			return c.JSON(http.StatusBadRequest, apiError)
		} else if ok {
			return c.NoContent(http.StatusOK)
		}
		return c.NoContent(http.StatusBadRequest)
	}
}
