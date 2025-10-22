package api

import "encoding/json"

type ShippingRequest struct {
	Weight        Weight        `json:"weight"`
	Shipper       Party         `json:"shipper"`
	Consignee     Party         `json:"consignee"`
	Dimensions    Dimensions    `json:"dimensions"`
	Packages      []Package     `json:"packages"`
	CustomsItems  []CustomsItem `json:"customsItems"`
	DeclaredValue Money         `json:"declaredValue"`
	ServiceType   string        `json:"serviceType"`
	SpecialNotes  string        `json:"specialNotes"`
	IsCOD         bool          `json:"isCod"`
	CODAmount     *Money        `json:"codAmount,omitempty"`
}

type Weight struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"` // "Grams", "KG"
}

type Party struct {
	Contact   Contact `json:"contact"`
	Address   Address `json:"address"`
	Reference string  `json:"reference"`
}

type Contact struct {
	Name        string `json:"name"`
	CompanyName string `json:"companyName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Mobile      string `json:"mobile"`
}

type Address struct {
	Line1       string `json:"line1"`
	Line2       string `json:"line2"`
	City        string `json:"city"`
	State       string `json:"state"`
	CountryCode string `json:"countryCode"`
	ZipCode     string `json:"zipCode"`
}

type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

type Package struct {
	Dimensions Dimensions `json:"dimensions"`
	Weight     float64    `json:"weight"`
	Quantity   int        `json:"quantity"`
	Value      float64    `json:"value"`
}

type CustomsItem struct {
	Description     string  `json:"description"`
	HSCode          string  `json:"hsCode"`
	Quantity        int     `json:"quantity"`
	Weight          float64 `json:"weight"`
	Value           float64 `json:"value"`
	CountryOfOrigin string  `json:"countryOfOrigin"`
}

type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type ShippingResponses struct {
	Responses []ShippingResponse `json:"responses"`
}

type ShippingResponse struct {
	Endpoint    string          `json:"enpodint"`
	RawResponse json.RawMessage `json:"rawReponse"`
	Error       string          `json:"error"`
}
