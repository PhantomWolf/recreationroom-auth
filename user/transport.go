package user

import (
	"github.com/PhantomWolf/recreationroom-auth/response"
	_ "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null"

	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// MakeHandler returns a handler for the user service.
func MakeHandler(serv Service, r *mux.Router) *mux.Router {
	opts := []kithttp.ServerOption{}
	createUserHandler := kithttp.NewServer(
		makeCreateUserEndpoint(serv),
		decodeCreateUserRequest,
		encodeResponse,
		opts...,
	)
	updateUserHandler := kithttp.NewServer(
		makeUpdateUserEndpoint(serv),
		decodeUpdateUserRequest,
		encodeResponse,
		opts...,
	)
	patchUserHandler := kithttp.NewServer(
		makePatchUserEndpoint(serv),
		decodePatchUserRequest,
		encodeResponse,
		opts...,
	)
	deleteUserHandler := kithttp.NewServer(
		makeDeleteUserEndpoint(serv),
		decodeDeleteUserRequest,
		encodeResponse,
		opts...,
	)
	getUserHandler := kithttp.NewServer(
		makeGetUserEndpoint(serv),
		decodeGetUserRequest,
		encodeResponse,
		opts...,
	)
	resetPasswordHandler := kithttp.NewServer(
		makeResetPasswordEndpoint(serv),
		decodeResetPasswordRequest,
		encodeResponse,
		opts...,
	)
	createPasswordHandler := kithttp.NewServer(
		makeCreatePasswordEndpoint(serv),
		decodeCreatePasswordRequest,
		encodeResponse,
		opts...,
	)
	updatePasswordHandler := kithttp.NewServer(
		makeUpdatePasswordEndpoint(serv),
		decodeUpdatePasswordEndpoint,
		encodeResponse,
		opts...,
	)

	// POST /users
	r.Handle("/users", createUserHandler).Methods("POST")
	// PUT /users/{id}
	r.Handle("/users/{id:[0-9]+}", updateUserHandler).Methods("PUT")
	// PATCH /users/{id}
	r.Handle("/users/{id:[0-9]+}", patchUserHandler).Methods("PATCH")
	// DELETE /users/{id}
	r.Handle("/users/{id:[0-9]+}", deleteUserHandler).Methods("DELETE")
	// GET /users/{id}
	r.Handle("/users/{id:[0-9]+}", getUserHandler).Methods("GET")
	// GET /password/reset
	r.Handle("/password/reset", resetPasswordHandler).Methods("GET")
	// POST /users/{id}/password
	r.Handle("/users/{id}/password", createPasswordHandler).Methods("POST")
	// PUT /users/{id}/password
	r.Handle("/users/{id}/password", updatePasswordHandler).Methods("PUT")
	return r
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, res interface{}) error {
	return json.NewEncoder(w).Encode(res.(*response.Response))
}

func decodeCreateUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Name     null.String `json:"name"`
		Password null.String `json:"password"`
		Email    null.String `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Debugf("[user/transport.go] Error decoding CreateUserRequest: %s\n", err.Error())
		return nil, err
	}
	return &createUserRequest{
		Name:     body.Name,
		Password: body.Password,
		Email:    body.Email,
	}, nil
}

func decodeUpdateUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Debugf("[user/transport.go] Invalid id %s: %s\n", vars["id"], err.Error())
		return nil, ErrInvalidRequest
	}

	var body struct {
		Name     null.String `json:"name"`
		Password null.String `json:"password"`
		Email    null.String `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Debugf("[user/transport.go] Error decoding UpdateUserRequest: %s\n", err.Error())
		return nil, err
	}
	return &updateUserRequest{
		ID:       id,
		Name:     body.Name,
		Password: body.Password,
		Email:    body.Email,
	}, nil
}

func decodePatchUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil || id <= 0 {
		log.Debugf("[user/transport.go] Invalid id %s: %s\n", vars["id"], err.Error())
		return nil, ErrInvalidRequest
	}

	var body struct {
		Name     null.String `json:"name"`
		Password null.String `json:"password"`
		Email    null.String `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Debugf("[user/transport.go] Error decoding PatchUserRequest: %s\n", err.Error())
		return nil, err
	}

	req := patchUserRequest{"id": id}
	if body.Name.Valid {
		req["name"] = body.Name.String
	}
	if body.Password.Valid {
		req["password"] = body.Password.String
	}
	if body.Email.Valid {
		req["email"] = body.Email.String
	}
	if len(req) < 2 {
		log.Debugf("[user/transport.go] Invalid request: %v\n", body)
		return nil, ErrInvalidRequest
	}
	return req, nil
}

func decodeDeleteUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil || id <= 0 {
		log.Debugf("[user/transport.go] Invalid id %s: %s\n", vars["id"], err.Error())
		return nil, ErrInvalidRequest
	}
	return &deleteUserRequest{ID: id}, nil
}

func decodeGetUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		log.Debugf("[user/transport.go] Invalid id %s: %s\n", vars["id"], err.Error())
		return nil, ErrInvalidRequest
	}
	return &getUserRequest{ID: id}, nil
}

// GET /password/reset
func decodeResetPasswordRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		NameOrEmail null.String `json:"name_or_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Debugf("[user/transport.go] Error decoding ResetPasswordRequest: %s\n", err.Error())
		return nil, err
	}
	return &resetPasswordRequest{NameOrEmail: body.NameOrEmail}, nil
}

// POST /users/<id>/password
func decodeCreatePasswordRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	var body struct {
		Token       null.String `json:"token"`
		NewPassword null.String `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &createPasswordRequest{ID: id, Token: body.Token, NewPassword: body.NewPassword}, nil
}

// PUT /users/<id>/password
func decodeUpdatePasswordEndpoint(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	var body struct {
		Password    null.String
		NewPassword null.String
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &updatePasswordRequest{ID: id, Password: body.Password, NewPassword: body.NewPassword}, nil
}
