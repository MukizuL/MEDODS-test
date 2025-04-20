package app

import (
	"MEDODS-test/internal/dto"
	"MEDODS-test/internal/errs"
	"MEDODS-test/internal/mailer"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type MyClaims struct {
	SessionID string `json:"sid,omitempty"`
	Type      string `json:"typ,omitempty"`
	Ip        string `json:"ip,omitempty"`
	jwt.RegisteredClaims
}

func (app *Application) CreateJWT(ctx context.Context, sub, ip string) (*dto.Tokens, error) {
	tokens, session, err := app.CreatePairOfTokens(sub, ip)
	if err != nil {
		app.logger.Error("Failed to create tokens", zap.Error(err))
		return nil, err
	}

	err = app.storage.SaveRefreshToken(ctx, tokens.RefreshToken, session)
	if err != nil {
		app.logger.Error("Failed to save refresh token", zap.Error(err))
		return nil, err
	}

	tokens.RefreshToken = base64.StdEncoding.EncodeToString([]byte(tokens.RefreshToken))

	return tokens, nil
}

func (app *Application) RefreshTokens(ctx context.Context, req *dto.RefreshRequest, ip string) (*dto.Tokens, error) {
	accessToken, err := jwt.ParseWithClaims(req.AccessToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(app.params.PrivateKey), nil
	})

	accessClaims, ok := accessToken.Claims.(*MyClaims)
	if !ok || !accessToken.Valid {
		return nil, fmt.Errorf("failed to parse an access token: %v", err)
	}

	rDec, err := base64.StdEncoding.DecodeString(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode a refresh token: %v", err)
	}

	refreshToken, err := jwt.ParseWithClaims(string(rDec), &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(app.params.PrivateKey), nil
	})

	refreshClaims, ok := refreshToken.Claims.(*MyClaims)
	if !ok || !refreshToken.Valid {
		return nil, fmt.Errorf("failed to parse a refresh token: %v", err)
	}

	// Check if tokens are from the same session
	if accessClaims.SessionID != refreshClaims.SessionID {
		return nil, fmt.Errorf("tokens are from different sessions")
	}

	// Check if user accessed from different ip
	if accessClaims.Ip != ip {
		err = mailer.SendMail(accessClaims.RegisteredClaims.Subject, "Your account accessed from different ip.")
		if err != nil {
			app.logger.Error("Failed to send email", zap.String("sub", accessClaims.RegisteredClaims.Subject))
		}
	}

	// Check if refresh token is not used
	err = app.storage.CheckoutRefreshToken(ctx, refreshToken.Raw, refreshClaims.SessionID)
	if err != nil {
		return nil, err
	}

	tokens, _, err := app.CreatePairOfTokens(refreshClaims.Subject, ip)
	if err != nil {
		app.logger.Error("Failed to create tokens", zap.Error(err))
		return nil, err
	}

	return tokens, nil
}

func (app *Application) CreatePairOfTokens(sub, ip string) (*dto.Tokens, string, error) {
	sessionID := uuid.New().String()
	claimsAccess := &MyClaims{
		SessionID: sessionID,
		Ip:        ip,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://" + app.params.Addr,
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	claimsRefresh := &MyClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claimsAccess)
	accessTokenSigned, err := accessToken.SignedString([]byte(app.params.PrivateKey))
	if err != nil {
		app.logger.Error("Error in CreateJWT", zap.Error(err))
		return nil, "", errs.ErrInternalServerError
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefresh)
	refreshTokenSigned, err := refreshToken.SignedString([]byte(app.params.PrivateKey))
	if err != nil {
		app.logger.Error("Error in CreateJWT", zap.Error(err))
		return nil, "", errs.ErrInternalServerError
	}

	return &dto.Tokens{
		AccessToken:  accessTokenSigned,
		RefreshToken: refreshTokenSigned,
	}, sessionID, nil
}
