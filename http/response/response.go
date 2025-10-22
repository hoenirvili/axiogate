package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	w http.ResponseWriter
}

func New(w http.ResponseWriter) Response {
	return Response{w: w}
}

func (r Response) OK(payload any) {
	r.w.WriteHeader(http.StatusOK)
	r.write(payload)
}

func (r Response) Created(payload any) {
	r.w.WriteHeader(http.StatusCreated)
	r.write(payload)
}

func (r Response) write(payload any) {
	if err := json.NewEncoder(r.w).
		Encode(payload); err != nil {
		panic(err)
	}
}

func (r Response) InternalServer(message string) {
	r.w.WriteHeader(http.StatusInternalServerError)
	r.write(&Error{Error: message})
}

type Error struct {
	Error string `json:"error"`
}

func (r Response) InternalServerf(format string, a ...any) {
	r.w.WriteHeader(http.StatusInternalServerError)
	r.write(&Error{Error: fmt.Sprintf(format, a...)})
}

func (r Response) NotFound(message string) {
	r.w.WriteHeader(http.StatusNotFound)
	r.write(&Error{Error: message})
}

func (r Response) NotFoundf(format string, a ...any) {
	r.w.WriteHeader(http.StatusNotFound)
	r.write(&Error{Error: fmt.Sprintf(format, a...)})
}

func (r Response) BadRequest(message string) {
	r.w.WriteHeader(http.StatusBadRequest)
	r.write(&Error{Error: message})
}

func (r Response) BadRequestf(format string, a ...any) {
	r.w.WriteHeader(http.StatusBadRequest)
	r.write(&Error{Error: fmt.Sprintf(format, a...)})
}
