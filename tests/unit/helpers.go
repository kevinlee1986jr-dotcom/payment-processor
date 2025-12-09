package payment_processor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
)

type validateError struct {
	Err error
}

func (ve *validateError) Error() string {
	return fmt.Sprintf("validate error: %v", ve.Err)
}

type paymentProcessor struct{}

func (p *paymentProcessor) validate(ctx context.Context, v *validator.Validate, data interface{}) error {
	err := v.Struct(data)
	if err != nil {
		return &validateError{err}
	}
	return nil
}

func (p *paymentProcessor) parseAmount(amount string) (float64, error) {
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount format")
	}
	var cents int64
	var err error
	if len(parts) == 2 {
		cents, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid cents value: %s", parts[1])
		}
	}
	dollars := int64(0)
	if parts[0] != "" {
		dollars, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid dollars value: %s", parts[0])
		}
	}
	return dollars*100 + cents, nil
}

func (p *paymentProcessor) generateToken(data jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString([]byte("secret_key"))
}

func (p *paymentProcessor) parseToken(token string) (jwt.MapClaims, error) {
	_, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return []byte("secret_key"), nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (p *paymentProcessor) httpError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	http.Error(w, message, code)
}

func (p *paymentProcessor) httpJsonError(ctx context.Context, w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	p.httpError(ctx, w, code, message)
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			p.httpError(ctx, w, code, err.Error(), nil)
		}
		w.Write(jsonData)
	}
}