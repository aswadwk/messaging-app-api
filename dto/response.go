package dto

// BaseResponse godoc
// @Schema
type BaseResponse struct {
	// @Example true
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"`
	Data    any      `json:"data"`
} //@name SuccessResponse

// ErrorResponse godoc
// @Schema
type ErrorResponse struct {
	// @Example false
	Success bool `json:"success"`
	// @Example "Incorrect API key"
	Message string `json:"message"`
	// @Example ["Invalid API key", "API key not found"]
	Errors []string `json:"errors,omitempty"`
	Data   any      `json:"data"`
} //@name ErrorResponse

type OccupationResponse struct {
	Occupation string `json:"occupation"`
}
