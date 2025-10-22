package b

import (
	"encoding/json"
	"fmt"

	"github.com/hoenirvili/axiogate/http/api"
)

type provider string

const Provider provider = "http://localhost:3031/v1/b"

type ProviderBRequest struct {
	Origin                       string                   `json:"Origin"`
	Destination                  string                   `json:"Destination"`
	ProductType                  string                   `json:"ProductType"`
	ServiceType                  string                   `json:"ServiceType"`
	CODAmount                    string                   `json:"CODAmount"`
	CODCurrency                  string                   `json:"CODCurrency"`
	SpecialInstruction           string                   `json:"SpecialInstruction"`
	Shipper                      string                   `json:"Shipper"`
	ShipperCPErson               string                   `json:"ShipperCPErson"`
	ShipperAddress1              string                   `json:"ShipperAddress1"`
	ShipperAddress2              string                   `json:"ShipperAddress2"`
	ShipperCity                  string                   `json:"ShipperCity"`
	ShipperEmail                 string                   `json:"ShipperEmail"`
	ShipperPhone                 string                   `json:"ShipperPhone"`
	ShipperMobile                string                   `json:"ShipperMobile"`
	ShipperRefNo                 string                   `json:"ShipperRefNo"`
	Consignee                    string                   `json:"Consignee"`
	ConsigneeCPerson             string                   `json:"ConsigneeCPerson"`
	ConsigneeAddress1            string                   `json:"ConsigneeAddress1"`
	ConsigneeAddress2            string                   `json:"ConsigneeAddress2"`
	ConsigneeCity                string                   `json:"ConsigneeCity"`
	ConsigneePhone               string                   `json:"ConsigneePhone"`
	ConsigneeMob                 string                   `json:"ConsigneeMob"`
	ConsigneeEmail               string                   `json:"ConsigneeEmail"`
	ConsigneeState               string                   `json:"ConsigneeState"`
	ConsigneeZipCode             string                   `json:"ConsigneeZipCode"`
	ValueOfShipment              float64                  `json:"ValueOfShipment"`
	ValueCurrency                string                   `json:"ValueCurrency"`
	GoodsDescription             string                   `json:"GoodsDescription"`
	NumberofPeices               int                      `json:"NumberofPeices"`
	Weight                       float64                  `json:"Weight"`
	PackageRequest               []PackageRequestB        `json:"PackageRequest"`
	ExportItemDeclarationRequest []ExportItemDeclarationB `json:"ExportItemDeclarationRequest"`
	UserName                     string                   `json:"UserName"`
	Password                     string                   `json:"Password"`
	AccountNo                    string                   `json:"AccountNo"`
}

type PackageRequestB struct {
	DimWidth      float64 `json:"DimWidth"`
	DimHeight     float64 `json:"DimHeight"`
	DimLength     float64 `json:"DimLength"`
	DimWeight     float64 `json:"DimWeight"`
	NoofPeices    int     `json:"NoofPeices"`
	ShipmentValue float64 `json:"ShipmentValue"`
}

type ExportItemDeclarationB struct {
	HSNCODE         string  `json:"HSNCODE"`
	ItemDesc        string  `json:"ItemDesc"`
	DimWeight       float64 `json:"DimWeight"`
	NoofPeices      int     `json:"NoofPeices"`
	ShipmentValue   float64 `json:"ShipmentValue"`
	CountryofOrigin string  `json:"CountryofOrigin"`
}

func (p provider) To() string {
	return string(p)
}

func (p provider) Payload(req *api.ShippingRequest) []byte {
	packages := make([]PackageRequestB, len(req.Packages))
	for i, pkg := range req.Packages {
		packages[i] = PackageRequestB{
			DimWidth:      pkg.Dimensions.Width,
			DimHeight:     pkg.Dimensions.Height,
			DimLength:     pkg.Dimensions.Length,
			DimWeight:     pkg.Weight,
			NoofPeices:    pkg.Quantity,
			ShipmentValue: pkg.Value,
		}
	}

	items := make([]ExportItemDeclarationB, len(req.CustomsItems))
	for i, item := range req.CustomsItems {
		items[i] = ExportItemDeclarationB{
			HSNCODE:         item.HSCode,
			ItemDesc:        item.Description,
			DimWeight:       item.Weight,
			NoofPeices:      item.Quantity,
			ShipmentValue:   item.Value,
			CountryofOrigin: item.CountryOfOrigin,
		}
	}

	codAmount := "0"
	codCurrency := "USD"
	if req.IsCOD && req.CODAmount != nil {
		codAmount = fmt.Sprintf("%.2f", req.CODAmount.Amount)
		codCurrency = req.CODAmount.Currency
	}
	goodsDescription := ""
	if len(req.CustomsItems) != 0 {
		goodsDescription = req.CustomsItems[0].Description
	}
	b, _ := json.Marshal(&ProviderBRequest{
		Origin:                       req.Shipper.Address.CountryCode,
		Destination:                  req.Consignee.Address.CountryCode,
		ProductType:                  "XPS",
		ServiceType:                  req.ServiceType,
		CODAmount:                    codAmount,
		CODCurrency:                  codCurrency,
		SpecialInstruction:           req.SpecialNotes,
		Shipper:                      req.Shipper.Contact.CompanyName,
		ShipperCPErson:               req.Shipper.Contact.Name,
		ShipperAddress1:              req.Shipper.Address.Line1,
		ShipperAddress2:              req.Shipper.Address.Line2,
		ShipperCity:                  req.Shipper.Address.City,
		ShipperEmail:                 req.Shipper.Contact.Email,
		ShipperPhone:                 req.Shipper.Contact.Phone,
		ShipperMobile:                req.Shipper.Contact.Mobile,
		ShipperRefNo:                 req.Shipper.Reference,
		Consignee:                    req.Consignee.Contact.CompanyName,
		ConsigneeCPerson:             req.Consignee.Contact.Name,
		ConsigneeAddress1:            req.Consignee.Address.Line1,
		ConsigneeAddress2:            req.Consignee.Address.Line2,
		ConsigneeCity:                req.Consignee.Address.City,
		ConsigneePhone:               req.Consignee.Contact.Phone,
		ConsigneeMob:                 req.Consignee.Contact.Mobile,
		ConsigneeEmail:               req.Consignee.Contact.Email,
		ConsigneeState:               req.Consignee.Address.State,
		ConsigneeZipCode:             req.Consignee.Address.ZipCode,
		ValueOfShipment:              req.DeclaredValue.Amount,
		ValueCurrency:                req.DeclaredValue.Currency,
		GoodsDescription:             goodsDescription,
		NumberofPeices:               len(req.Packages),
		Weight:                       req.Weight.Value / 1000, // Convert to KG
		PackageRequest:               packages,
		ExportItemDeclarationRequest: items,
		UserName:                     "123", // From config
		Password:                     "123",
		AccountNo:                    "123",
	})
	return b
}
