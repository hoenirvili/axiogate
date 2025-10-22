package a

import (
	"encoding/json"
	"fmt"

	"github.com/hoenirvili/axiogate/http/api"
)

type provider string

const Provider provider = "http://localhost:3030/v1/a"

type ProviderARequest struct {
	Weight              WeightA               `json:"weight"`
	Shipper             PartyA                `json:"shipper"`
	Consignee           PartyA                `json:"consignee"`
	Dimensions          DimensionsA           `json:"dimensions"`
	Account             AccountA              `json:"account"`
	ProductCode         string                `json:"productCode"`
	ServiceType         string                `json:"serviceType"`
	PrintType           string                `json:"printType"`
	IsInsured           bool                  `json:"isInsured"`
	CustomsDeclarations []CustomsDeclarationA `json:"customsDeclarations"`
	DeclaredValue       MoneyA                `json:"declaredValue"`
	NumberOfPieces      int                   `json:"numberOfPieces"`
	ReferenceNumber1    string                `json:"referenceNumber1"`
	SpecialNotes        string                `json:"specialNotes"`
	Remarks             string                `json:"remarks"`
	DeliveryType        string                `json:"deliveryType"`
	ContentType         string                `json:"contentType"`
	IsCod               bool                  `json:"isCod"`
}

type WeightA struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type PartyA struct {
	Contact      ContactA `json:"contact"`
	Address      AddressA `json:"address"`
	ReferenceNo1 string   `json:"referenceNo1"`
}

type ContactA struct {
	Name         string `json:"name"`
	MobileNumber string `json:"mobileNumber"`
	PhoneNumber  string `json:"phoneNumber"`
	EmailAddress string `json:"emailAddress"`
	CompanyName  string `json:"companyName"`
}

type AddressA struct {
	Line1       string `json:"line1"`
	City        string `json:"city"`
	CountryCode string `json:"countryCode"`
	ZipCode     string `json:"zipCode"`
}

type DimensionsA struct {
	Length float64 `json:"length"`
	Height float64 `json:"height"`
	Width  float64 `json:"width"`
	Unit   string  `json:"unit"`
}

type AccountA struct {
	Number int `json:"number"`
}

type CustomsDeclarationA struct {
	Reference       string      `json:"reference"`
	Description     string      `json:"description"`
	CountryOfOrigin string      `json:"countryOfOrigin"`
	Weight          float64     `json:"weight"`
	Dimensions      DimensionsA `json:"dimensions"`
	Quantity        int         `json:"quantity"`
	HsCode          string      `json:"hsCode"`
	Value           float64     `json:"value"`
}

type MoneyA struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (p provider) To() string {
	return string(p)
}

func (p provider) Payload(req *api.ShippingRequest) []byte {
	customs := make([]CustomsDeclarationA, len(req.CustomsItems))
	for i, item := range req.CustomsItems {
		customs[i] = CustomsDeclarationA{
			Reference:       fmt.Sprintf("Item-%d", i+1),
			Description:     item.Description,
			CountryOfOrigin: item.CountryOfOrigin,
			Weight:          item.Weight,
			Dimensions: DimensionsA{
				Length: req.Dimensions.Length,
				Width:  req.Dimensions.Width,
				Height: req.Dimensions.Height,
				Unit:   req.Dimensions.Unit,
			},
			Quantity: item.Quantity,
			HsCode:   item.HSCode,
			Value:    item.Value,
		}
	}

	b, _ := json.Marshal(&ProviderARequest{
		Weight: WeightA{
			Value: req.Weight.Value,
			Unit:  req.Weight.Unit,
		},
		Shipper: PartyA{
			Contact: ContactA{
				Name:         req.Shipper.Contact.Name,
				MobileNumber: req.Shipper.Contact.Mobile,
				PhoneNumber:  req.Shipper.Contact.Phone,
				EmailAddress: req.Shipper.Contact.Email,
				CompanyName:  req.Shipper.Contact.CompanyName,
			},
			Address: AddressA{
				Line1:       req.Shipper.Address.Line1,
				City:        req.Shipper.Address.City,
				CountryCode: req.Shipper.Address.CountryCode,
				ZipCode:     req.Shipper.Address.ZipCode,
			},
			ReferenceNo1: req.Shipper.Reference,
		},
		Consignee: PartyA{
			Contact: ContactA{
				Name:         req.Consignee.Contact.Name,
				MobileNumber: req.Consignee.Contact.Mobile,
				PhoneNumber:  req.Consignee.Contact.Phone,
				EmailAddress: req.Consignee.Contact.Email,
				CompanyName:  req.Consignee.Contact.CompanyName,
			},
			Address: AddressA{
				Line1:       req.Consignee.Address.Line1,
				City:        req.Consignee.Address.City,
				CountryCode: req.Consignee.Address.CountryCode,
				ZipCode:     req.Consignee.Address.ZipCode,
			},
			ReferenceNo1: req.Consignee.Reference,
		},
		Dimensions: DimensionsA{
			Length: req.Dimensions.Length,
			Width:  req.Dimensions.Width,
			Height: req.Dimensions.Height,
			Unit:   req.Dimensions.Unit,
		},
		Account: AccountA{
			Number: 123, // Should come from config
		},
		ProductCode:         "International",
		ServiceType:         req.ServiceType,
		PrintType:           "AWBOnly",
		IsInsured:           true,
		CustomsDeclarations: customs,
		DeclaredValue: MoneyA{
			Amount:   req.DeclaredValue.Amount,
			Currency: req.DeclaredValue.Currency,
		},
		NumberOfPieces: len(req.Packages),
		SpecialNotes:   req.SpecialNotes,
		DeliveryType:   "DoorToDoor",
		ContentType:    "NonDocument",
		IsCod:          req.IsCOD,
	})
	return b
}
