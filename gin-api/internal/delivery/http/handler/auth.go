package handler

import (
	"context"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type AuthHandler struct {
	useCase *usecase.AuthUseCase
}

func NewAuthHandler(useCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

func (h *AuthHandler) Register(
	ctx context.Context,
	request api.RegisterRequestObject,
) (api.RegisterResponseObject, error) {
	result, err := h.useCase.Register(
		ctx,
		string(request.Body.Email),
		request.Body.Password,
	)
	if err != nil {
		return nil, err
	}

	return api.Register202JSONResponse{
		VerificationId: result.VerificationID,
		ExpiresIn:      result.ExpiresIn,
	}, nil
}

func (h *AuthHandler) ResendRegistrationVerification(
	ctx context.Context,
	request api.ResendRegistrationVerificationRequestObject,
) (api.ResendRegistrationVerificationResponseObject, error) {
	result, err := h.useCase.ResendRegistrationVerification(
		ctx,
		request.Body.VerificationId,
	)
	if err != nil {
		return nil, err
	}

	return api.ResendRegistrationVerification202JSONResponse{
		VerificationId: result.VerificationID,
		ExpiresIn:      result.ExpiresIn,
		RetryAfter:     result.RetryAfter,
	}, nil
}

func (h *AuthHandler) VerifyRegistration(
	ctx context.Context,
	request api.VerifyRegistrationRequestObject,
) (api.VerifyRegistrationResponseObject, error) {
	user, err := h.useCase.VerifyRegistration(
		ctx,
		request.Body.VerificationId,
		request.Body.Code,
	)
	if err != nil {
		return nil, err
	}

	return api.VerifyRegistration201JSONResponse{
		Id:    user.ID,
		Email: openapi_types.Email(user.Email),
	}, nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	request api.LoginRequestObject,
) (api.LoginResponseObject, error) {
	result, err := h.useCase.Login(
		ctx,
		string(request.Body.Email),
		request.Body.Password,
	)
	if err != nil {
		return nil, err
	}

	return api.Login200JSONResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (h *AuthHandler) Refresh(
	ctx context.Context,
	request api.RefreshRequestObject,
) (api.RefreshResponseObject, error) {
	result, err := h.useCase.Refresh(
		ctx,
		request.Body.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	return api.Refresh200JSONResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (h *AuthHandler) Logout(
	ctx context.Context,
	request api.LogoutRequestObject,
) (api.LogoutResponseObject, error) {
	if err := h.useCase.Logout(ctx, request.Body.RefreshToken); err != nil {
		return nil, err
	}

	return api.Logout204Response{}, nil
}
