package handler

import (
	"context"
	"log"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type AuthHandler struct {
	useCase *usecase.AuthUseCase
}

func NewAuthHandler(
	useCase *usecase.AuthUseCase,
) *AuthHandler {
	return &AuthHandler{
		useCase: useCase,
	}
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
		log.Printf(err.Error())
		return nil, err
	}

	return api.VerifyRegistration201JSONResponse{
		Id:    user.ID,
		Email: openapi_types.Email(user.Email),
	}, nil
}
