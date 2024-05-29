package main

import (
	"context"
	"flowgest/auth/cmd/models"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type PasskeyUser interface {
	webauthn.User
	AddCredential(*webauthn.Credential)
	UpdateCredential(*webauthn.Credential)
}

type PasskeyStore interface {
	GetUser(userName string) PasskeyUser
	SaveUser(PasskeyUser)
	GetSession(token string) webauthn.SessionData
	SaveSession(token string, data webauthn.SessionData)
	DeleteSession(token string)
}

var (
	webAuthn  *webauthn.WebAuthn
	err       error
	datastore PasskeyStore
)

func main() {
	proto := getEnv("PROTO", "http")
	host := getEnv("HOST", "localhost")
	port := getEnv("PORT", ":8080")
	origin := fmt.Sprintf("%s://%s%s", proto, host, port)

	wconfig := &webauthn.Config{
		RPDisplayName: "Go Webauthn",    // Display Name for your site
		RPID:          host,             // Generally the FQDN for your site
		RPOrigins:     []string{origin}, // The origin URLs allowed for WebAuthn
	}
	if webAuthn, err = webauthn.New(wconfig); err != nil {
		fmt.Printf("[FATA] %s", err.Error())
		os.Exit(1)
	}

	datastore = NewInMem()

	e := echo.New()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	//TODO: see alternatives of logging
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogProtocol: true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				// logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
				// 	slog.String("uri", v.URI),
				// 	slog.Int("status", v.Status),
				// )

				logger.InfoContext(c.Request().Context(), "", slog.Group("Request",
					slog.String("uri", v.URI),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
					slog.String("Protocol", v.Protocol),
				))
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	
	e.Static("/", "web")
	e.File("/", "web/index.html")
	e.POST("/api/passkey/registerStart", func(c echo.Context) error {
		fmt.Printf("[INFO] begin registration ----------------------\\")

		userdto := new(models.UserDTO)
		err := c.Bind(&userdto)
		if err != nil {
			fmt.Printf("[ERROR] can't get username: %s", err.Error())
			panic(err)
		}

		user := datastore.GetUser(userdto.Username) // Find or create the new user

		options, session, err := webAuthn.BeginRegistration(user)
		if err != nil {
			msg := fmt.Sprintf("can't begin registration: %s", err.Error())
			fmt.Printf("[ERROR] %s", msg)
			return c.JSON(http.StatusBadRequest, msg)
		}

		// Make a session key and store the sessionData values
		token := uuid.New().String()
		datastore.SaveSession(token, *session)

		c.Response().Header().Set("Session-Key", token)
		return c.JSON(http.StatusOK, options)
	})
	e.POST("/api/passkey/registerFinish", func(c echo.Context) error {
		// Get the session key from the header
		token := c.Request().Header.Get("Session-Key")
		// Get the session data stored from the function above
		session := datastore.GetSession(token) // FIXME: cover invalid session

		// In out example username == userID, but in real world it should be different
		user := datastore.GetUser(string(session.UserID)) // Get the user

		credential, err := webAuthn.FinishRegistration(user, session, c.Request())
		if err != nil {
			msg := fmt.Sprintf("can't finish registration: %s", err.Error())
			fmt.Printf("[ERROR] %s", msg)
			return c.JSON(http.StatusBadRequest, msg)
		}

		// If creation was successful, store the credential object
		user.AddCredential(credential)
		datastore.SaveUser(user)
		// Delete the session data
		datastore.DeleteSession(token)

		fmt.Printf("[INFO] finish registration ----------------------/")
		return c.JSON(http.StatusOK, "Registration Success") // Handle next steps
	})
	e.POST("/api/passkey/loginStart", func(c echo.Context) error {
		fmt.Printf("[INFO] begin login ----------------------\\")

		userdto := new(models.UserDTO)
		err := c.Bind(&userdto)
		if err != nil {
			fmt.Printf("[ERROR]can't get user name: %s", err.Error())
			panic(err)
		}

		user := datastore.GetUser(userdto.Username) // Find the user

		options, session, err := webAuthn.BeginLogin(user)
		if err != nil {
			msg := fmt.Sprintf("can't begin login: %s", err.Error())
			fmt.Printf("[ERROR] %s", msg)
			return c.JSON(http.StatusBadRequest, msg)
		}

		// Make a session key and store the sessionData values
		token := uuid.New().String()
		datastore.SaveSession(token, *session)

		c.Response().Header().Set("Session-Key", token)
		return c.JSON(http.StatusOK, options)
	})
	e.POST("/api/passkey/loginFinish", func(c echo.Context) error {
		// Get the session key from the header
		token := c.Request().Header.Get("Session-Key")
		// Get the session data stored from the function above
		session := datastore.GetSession(token) // FIXME: cover invalid session

		// In out example username == userID, but in real world it should be different
		user := datastore.GetUser(string(session.UserID)) // Get the user

		credential, err := webAuthn.FinishLogin(user, session, c.Request())
		if err != nil {
			fmt.Printf("[ERROR] can't finish login %s", err.Error())
			panic(err)
		}

		// Handle credential.Authenticator.CloneWarning
		if credential.Authenticator.CloneWarning {
			fmt.Printf("[WARN] can't finish login: %s", "CloneWarning")
		}

		// If login was successful, update the credential object
		user.UpdateCredential(credential)
		datastore.SaveUser(user)
		// Delete the session data
		datastore.DeleteSession(token)

		fmt.Printf("[INFO] finish login ----------------------/")
		return c.JSON(http.StatusOK, "Login Success")
	})

	e.Logger.Fatal(e.Start(":3000"))
}

// getEnv is a helper function to get the environment variable
func getEnv(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}
