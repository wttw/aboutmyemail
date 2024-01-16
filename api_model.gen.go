// Package aboutmyemail provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.0.0 DO NOT EDIT.
package aboutmyemail

import (
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	BearerAuthScopes = "BearerAuth.Scopes"
)

// N400Error defines model for 400Error.
type N400Error struct {
	Message string `json:"message"`
}

// N500Error defines model for 500Error.
type N500Error struct {
	Message string `json:"message"`
}

// StatusResult defines model for StatusResult.
type StatusResult struct {
	Id       string    `json:"id"`
	Messages *[]string `json:"messages,omitempty"`
	Token    *string   `json:"token,omitempty"`
	Url      *string   `json:"url,omitempty"`
}

// Submit defines model for Submit.
type Submit struct {
	FinishedUrl *string `json:"finishedUrl,omitempty"`
	From        string  `json:"from"`
	Ip          string  `json:"ip"`
	Payload     string  `json:"payload"`
	ProgressUrl *string `json:"progressUrl,omitempty"`
	Smtputf8    *bool   `json:"smtputf8,omitempty"`
	To          string  `json:"to"`
	Token       *string `json:"token,omitempty"`
}

// SubmitForm defines model for SubmitForm.
type SubmitForm struct {
	FinishedUrl *string            `json:"finishedUrl,omitempty"`
	From        string             `json:"from"`
	Ip          string             `json:"ip"`
	Payload     openapi_types.File `json:"payload"`
	ProgressUrl *string            `json:"progressUrl,omitempty"`
	Smtputf8    *bool              `json:"smtputf8,omitempty"`
	To          string             `json:"to"`
	Token       *string            `json:"token,omitempty"`
}

// SubmitSuccess defines model for SubmitSuccess.
type SubmitSuccess struct {
	Id string `json:"id"`
}

// EmailJSONRequestBody defines body for Email for application/json ContentType.
type EmailJSONRequestBody = Submit

// EmailMultipartRequestBody defines body for Email for multipart/form-data ContentType.
type EmailMultipartRequestBody = SubmitForm