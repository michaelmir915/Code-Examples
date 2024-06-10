//go:build cgo && windows
// +build cgo,windows

// NOTE: Many test cases were left in as per client request
package main

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lmyclibrary
*/
import (
	"C" //Needed for Csharp integration
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	_ "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/skratchdot/open-golang/open"

	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
)

// Structs created for parts
type Categories struct {
	Name        string        `json:"Name"`
	ID          sql.NullInt64 `json:"ID"`
	Description string        `json:"Description"`
}

// Meant when updating a new part
type updatePart struct {
	ID                       int             `json:"ID"`
	Description              sql.NullString  `json:"Description"`
	Manufacturer             sql.NullString  `json:"Manufacturer"`
	Manufacturer_Part_Number sql.NullString  `json:"Manu_Part_Number"`
	Supplier_Part_Number     sql.NullString  `json:"Supplier_Part_Number"`
	Part_Category            sql.NullString  `json:"Category"`
	Cost_1pc                 sql.NullFloat64 `json:"cost1pc"`
	Cost_100pc               sql.NullFloat64 `json:"cost100pc"`
	Cost_1000pc              sql.NullFloat64 `json:"cost1000pc"`
	PrimaryVendorStock       sql.NullInt64   `json:"primary_stock"`
	RoHS                     sql.NullBool    `json:"RoHS Verified"`
}

// Temporarily used when making a new part
type newPart struct {
	Supplier_Part_Number string `json:"Supplier_Part_Number"`
}

// Struct used when searching on mouser
type mouserParams struct {
	mouserPartNumber  string `json:"Part Number"`
	partSearchOptions string `json:"Search Options"`
}

// Initialize a similar struct for Digi-key
type digiKeyParams struct {
	digiKeyPartNumber string `json:"Part Number"`
}

// Mouser JSON Response
type mouserResponse struct {
	Errors        []any `json:"Errors"`
	SearchResults struct {
		NumberOfResult int `json:"NumberOfResult"`
		Parts          []struct {
			Availability           string `json:"Availability"`
			DataSheetURL           string `json:"DataSheetUrl"`
			Description            string `json:"Description"`
			FactoryStock           string `json:"FactoryStock"`
			ImagePath              string `json:"ImagePath"`
			Category               string `json:"Category"`
			LeadTime               string `json:"LeadTime"`
			LifecycleStatus        any    `json:"LifecycleStatus"`
			Manufacturer           string `json:"Manufacturer"`
			ManufacturerPartNumber string `json:"ManufacturerPartNumber"`
			Min                    string `json:"Min"`
			Mult                   string `json:"Mult"`
			MouserPartNumber       string `json:"MouserPartNumber"`
			ProductAttributes      []struct {
				AttributeName  string `json:"AttributeName"`
				AttributeValue string `json:"AttributeValue"`
			} `json:"ProductAttributes"`
			PriceBreaks []struct {
				Quantity int    `json:"Quantity"`
				Price    string `json:"Price"`
				Currency string `json:"Currency"`
			} `json:"PriceBreaks"`
			AlternatePackagings  any    `json:"AlternatePackagings"`
			ProductDetailURL     string `json:"ProductDetailUrl"`
			Reeling              bool   `json:"Reeling"`
			ROHSStatus           string `json:"ROHSStatus"`
			SuggestedReplacement string `json:"SuggestedReplacement"`
			MultiSimBlue         int    `json:"MultiSimBlue"`
			AvailabilityInStock  string `json:"AvailabilityInStock"`
			AvailabilityOnOrder  []any  `json:"AvailabilityOnOrder"`
			InfoMessages         []any  `json:"InfoMessages"`
			RestrictionMessage   string `json:RestrictionMessage`
			ProductCompliance    []struct {
				ComplianceName  string `json:"ComplianceName"`
				ComplianceValue string `json:"ComplianceValue"`
			} `json:"ProductCompliance"`
		} `json:"Parts"`
	} `json:"SearchResults"`
}

// Digi-key Refresh Token https://api.digikey.com/v1/oauth2/token
type refreshToken struct {
	accessToken        string `json:"access_token"`
	refreshToken       string `json:"refresh_token"`
	expiresIn          int8   `json:"expires_in"`
	refreshTokenExpire int16  `json:"refresh_token_expires_in"`
	tokenType          string `json:"token_type"`
}

// Digi-key JSON Response
type digiKeyResponse struct {
	MyPricing  []any `json:"MyPricing"`
	Obsolete   bool  `json:"Obsolete"`
	MediaLinks []struct {
		MediaType  string `json:"MediaType"`
		Title      string `json:"Title"`
		SmallPhoto string `json:"SmallPhoto"`
		Thumbnail  string `json:"Thumbnail"`
		URL        string `json:"Url"`
	} `json:"MediaLinks"`
	StandardPackage int `json:"StandardPackage"`
	LimitedTaxonomy struct {
		Children []struct {
			Children        []any  `json:"Children"`
			ProductCount    int    `json:"ProductCount"`
			NewProductCount int    `json:"NewProductCount"`
			ParameterID     int    `json:"ParameterId"`
			ValueID         string `json:"ValueId"`
			Parameter       string `json:"Parameter"`
			Value           string `json:"Value"`
		} `json:"Children"`
		ProductCount    int    `json:"ProductCount"`
		NewProductCount int    `json:"NewProductCount"`
		ParameterID     int    `json:"ParameterId"`
		ValueID         string `json:"ValueId"`
		Parameter       string `json:"Parameter"`
		Value           string `json:"Value"`
	} `json:"LimitedTaxonomy"`
	Kits             []any `json:"Kits"`
	KitContents      []any `json:"KitContents"`
	MatingProducts   []any `json:"MatingProducts"`
	SearchLocaleUsed struct {
		Site          string `json:"Site"`
		Language      string `json:"Language"`
		Currency      string `json:"Currency"`
		ShipToCountry string `json:"ShipToCountry"`
	} `json:"SearchLocaleUsed"`
	AssociatedProducts []struct {
		ProductURL             string `json:"ProductUrl"`
		ManufacturerPartNumber string `json:"ManufacturerPartNumber"`
		MinimumOrderQuantity   int    `json:"MinimumOrderQuantity"`
		NonStock               bool   `json:"NonStock"`
		Packaging              struct {
			ParameterID int    `json:"ParameterId"`
			ValueID     string `json:"ValueId"`
			Parameter   string `json:"Parameter"`
			Value       string `json:"Value"`
		} `json:"Packaging"`
		QuantityAvailable  int     `json:"QuantityAvailable"`
		DigiKeyPartNumber  string  `json:"DigiKeyPartNumber"`
		ProductDescription string  `json:"ProductDescription"`
		UnitPrice          float64 `json:"UnitPrice"`
		Manufacturer       struct {
			ParameterID int    `json:"ParameterId"`
			ValueID     string `json:"ValueId"`
			Parameter   string `json:"Parameter"`
			Value       string `json:"Value"`
		} `json:"Manufacturer"`
		ManufacturerPublicQuantity int    `json:"ManufacturerPublicQuantity"`
		QuantityOnOrder            int    `json:"QuantityOnOrder"`
		MaxQuantityForDistribution int    `json:"MaxQuantityForDistribution"`
		BackOrderNotAllowed        bool   `json:"BackOrderNotAllowed"`
		DKPlusRestriction          bool   `json:"DKPlusRestriction"`
		Marketplace                bool   `json:"Marketplace"`
		SupplierDirectShip         bool   `json:"SupplierDirectShip"`
		PimProductName             string `json:"PimProductName"`
		Supplier                   string `json:"Supplier"`
		SupplierID                 int    `json:"SupplierId"`
		IsNcnr                     bool   `json:"IsNcnr"`
	} `json:"AssociatedProducts"`
	ForUseWithProducts []struct {
		ProductURL             string `json:"ProductUrl"`
		ManufacturerPartNumber string `json:"ManufacturerPartNumber"`
		MinimumOrderQuantity   int    `json:"MinimumOrderQuantity"`
		NonStock               bool   `json:"NonStock"`
		Packaging              struct {
			ParameterID int    `json:"ParameterId"`
			ValueID     string `json:"ValueId"`
			Parameter   string `json:"Parameter"`
			Value       string `json:"Value"`
		} `json:"Packaging"`
		QuantityAvailable  int     `json:"QuantityAvailable"`
		DigiKeyPartNumber  string  `json:"DigiKeyPartNumber"`
		ProductDescription string  `json:"ProductDescription"`
		UnitPrice          float64 `json:"UnitPrice"`
		Manufacturer       struct {
			ParameterID int    `json:"ParameterId"`
			ValueID     string `json:"ValueId"`
			Parameter   string `json:"Parameter"`
			Value       string `json:"Value"`
		} `json:"Manufacturer"`
		ManufacturerPublicQuantity int    `json:"ManufacturerPublicQuantity"`
		QuantityOnOrder            int    `json:"QuantityOnOrder"`
		MaxQuantityForDistribution int    `json:"MaxQuantityForDistribution"`
		BackOrderNotAllowed        bool   `json:"BackOrderNotAllowed"`
		DKPlusRestriction          bool   `json:"DKPlusRestriction"`
		Marketplace                bool   `json:"Marketplace"`
		SupplierDirectShip         bool   `json:"SupplierDirectShip"`
		PimProductName             string `json:"PimProductName"`
		Supplier                   string `json:"Supplier"`
		SupplierID                 int    `json:"SupplierId"`
		IsNcnr                     bool   `json:"IsNcnr"`
	} `json:"ForUseWithProducts"`
	RohsSubs           []any   `json:"RohsSubs"`
	SuggestedSubs      []any   `json:"SuggestedSubs"`
	AdditionalValueFee float64 `json:"AdditionalValueFee"`
	ReachEffectiveDate string  `json:"ReachEffectiveDate"`
	ShippingInfo       string  `json:"ShippingInfo"`
	StandardPricing    []struct {
		BreakQuantity int     `json:"BreakQuantity"`
		UnitPrice     float64 `json:"UnitPrice"`
		TotalPrice    float64 `json:"TotalPrice"`
	} `json:"StandardPricing"`
	RoHSStatus string `json:"RoHSStatus"`
	LeadStatus string `json:"LeadStatus"`
	Parameters []struct {
		ParameterID int    `json:"ParameterId"`
		ValueID     string `json:"ValueId"`
		Parameter   string `json:"Parameter"`
		Value       string `json:"Value"`
	} `json:"Parameters"`
	ProductURL       string `json:"ProductUrl"`
	PrimaryDatasheet string `json:"PrimaryDatasheet"`
	PrimaryPhoto     string `json:"PrimaryPhoto"`
	PrimaryVideo     string `json:"PrimaryVideo"`
	Series           struct {
		ParameterID int    `json:"ParameterId"`
		ValueID     string `json:"ValueId"`
		Parameter   string `json:"Parameter"`
		Value       string `json:"Value"`
	} `json:"Series"`
	ManufacturerLeadWeeks    string `json:"ManufacturerLeadWeeks"`
	ManufacturerPageURL      string `json:"ManufacturerPageUrl"`
	ProductStatus            string `json:"ProductStatus"`
	AlternatePackaging       []any  `json:"AlternatePackaging"`
	DetailedDescription      string `json:"DetailedDescription"`
	ReachStatus              string `json:"ReachStatus"`
	ExportControlClassNumber string `json:"ExportControlClassNumber"`
	HTSUSCode                string `json:"HTSUSCode"`
	TariffDescription        string `json:"TariffDescription"`
	MoistureSensitivityLevel string `json:"MoistureSensitivityLevel"`
	Family                   struct {
		ParameterID int    `json:"ParameterId"`
		ValueID     string `json:"ValueId"`
		Parameter   string `json:"Parameter"`
		Value       string `json:"Value"`
	} `json:"Family"`
	Category struct {
		ParameterID int    `json:"ParameterId"`
		ValueID     string `json:"ValueId"`
		Parameter   string `json:"Parameter"`
		Value       string `json:"Value"`
	} `json:"Category"`
	ManufacturerPartNumber string `json:"ManufacturerPartNumber"`
	MinimumOrderQuantity   int    `json:"MinimumOrderQuantity"`
	NonStock               bool   `json:"NonStock"`
	Packaging              struct {
		ParameterID int    `json:"ParameterId"`
		ValueID     string `json:"ValueId"`
		Parameter   string `json:"Parameter"`
		Value       string `json:"Value"`
	} `json:"Packaging"`
	QuantityAvailable  int     `json:"QuantityAvailable"`
	DigiKeyPartNumber  string  `json:"DigiKeyPartNumber"`
	ProductDescription string  `json:"ProductDescription"`
	UnitPrice          float64 `json:"UnitPrice"`
	Manufacturer       struct {
		ParameterID int    `json:"ParameterId"`
		ValueID     string `json:"ValueId"`
		Parameter   string `json:"Parameter"`
		Value       string `json:"Value"`
	} `json:"Manufacturer"`
	ManufacturerPublicQuantity int    `json:"ManufacturerPublicQuantity"`
	QuantityOnOrder            int    `json:"QuantityOnOrder"`
	MaxQuantityForDistribution int    `json:"MaxQuantityForDistribution"`
	BackOrderNotAllowed        bool   `json:"BackOrderNotAllowed"`
	DKPlusRestriction          bool   `json:"DKPlusRestriction"`
	Marketplace                bool   `json:"Marketplace"`
	SupplierDirectShip         bool   `json:"SupplierDirectShip"`
	PimProductName             string `json:"PimProductName"`
	Supplier                   string `json:"Supplier"`
	SupplierID                 int    `json:"SupplierId"`
	IsNcnr                     bool   `json:"IsNcnr"`
}

// Setup a connection to the database
func dbConnection() *sql.DB {
	db, err := sql.Open("mysql", "root:Password!@tcp(localhost:3306)/lusher engineering parts database")
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Adding a new part to the database, will point to the proper function depending on supplier
func addNewPart(partSupplier string, newPartNumber string) {
	var partToAdd newPart
	partToAdd.Supplier_Part_Number = newPartNumber
	if partSupplier == "Digi-Key" {
		newPart_Digikey(partToAdd)
	} else if partSupplier == "Mouser Electronics" {
		newPart_Mouser(partToAdd)
	}
}

// Digikey update parts
func update_Digikey(partsList []int) {
	db := dbConnection()
	// partsList := (gatherDatabase(db))
	genTokens()
	for i := range partsList { //i in range parts list, if only one part to update it will only do it once
		var partToAdd updatePart
		partToAdd.ID = partsList[i]
		err := godotenv.Load("../FrontEnd/.env") //using .env file for easier access
		if err != nil {
			log.Fatal("Error loading .env file") //error handling
		}
		err = godotenv.Load("../FrontEnd/tokens.env") //using .env file for easier access
		if err != nil {
			log.Fatal("Error loading tokens.env file") //error handling
		}
		var supplierPartNumber string
		//get the old information from the list
		err = db.QueryRow("SELECT `Supplier Part Number 1` FROM `lusher engineering parts database`.electronics_parts WHERE `ID`=?", partsList[i]).Scan(&supplierPartNumber)
		if err != nil {
			fmt.Println("Could not query Database")
			log.Fatal(err)
		}

		//Retreive the part number
		partToAdd.Supplier_Part_Number.String = supplierPartNumber
		partToAdd.Supplier_Part_Number.Valid = true
		digiKeyPartNumber := supplierPartNumber
		partSearchOptions := "string"

		//Access with api Key
		clientID := os.Getenv("DIGI_ID")
		accessToken := "Bearer " + os.Getenv("ACCESS_TOKEN")

		partURL := "https://api.digikey.com/Search/v3/Products/" + partToAdd.Supplier_Part_Number.String

		// //For testing purposes
		// fmt.Println(clientID)
		// fmt.Println(accessToken)
		// //Print URL for testing Purposes, TODO: remove or comment with integration
		// fmt.Println("URL:>", partURL)

		// create a map to hold the request data
		requestData := map[string]interface{}{
			"SearchByPartRequest": map[string]string{
				"digiKeyPartNumber": digiKeyPartNumber,
				"partSearchOptions": partSearchOptions,
			},
		}

		// convert the map to a JSON byte array
		jsonStr, err := json.Marshal(requestData)
		req, err := http.NewRequest("GET", partURL, bytes.NewBuffer(jsonStr))
		req.Header.Set("accept", "application/json")
		req.Header.Set("Authorization", accessToken)
		req.Header.Set("X-DIGIKEY-Client-Id", clientID)
		//reponse for post requestion
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		//Response should always be 200, TODO: error handling
		fmt.Println("response Status:", resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		var results1 digiKeyResponse
		///Unmarshal results to more readable data
		err = json.Unmarshal(body, &results1)
		if err != nil {
			log.Fatal(err)
		}

		// // pretty print json on terminal for testing
		// data, err := json.MarshalIndent(results1, "", "  ")
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println(string(data))

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Received HTTP status: %s, likely could not find the part.\n", resp.Status)
		} else {
			fmt.Println("Response is 200 OK")
			partToAdd.Description.String = results1.DetailedDescription
			partToAdd.Description.Valid = true
			partToAdd.Manufacturer.String = results1.Manufacturer.Value
			partToAdd.Manufacturer.Valid = true
			partToAdd.Manufacturer_Part_Number.String = results1.ManufacturerPartNumber
			partToAdd.Manufacturer_Part_Number.Valid = true
			partToAdd.Supplier_Part_Number.String = results1.DigiKeyPartNumber
			partToAdd.Part_Category.String = results1.Category.Value
			partToAdd.Part_Category.Valid = true
			if results1.RoHSStatus == "ROHS3 Compliant" {
				partToAdd.RoHS.Bool = true
			} else {
				partToAdd.RoHS.Bool = false
			}
			partToAdd.RoHS.Valid = true
			partToAdd.PrimaryVendorStock.Int64 = int64(results1.QuantityAvailable)
			partToAdd.PrimaryVendorStock.Valid = true
			for _, priceInfo := range results1.StandardPricing {
				if priceInfo.BreakQuantity == 1 {
					partToAdd.Cost_1pc.Float64 = priceInfo.UnitPrice
					partToAdd.Cost_1pc.Valid = true
				} else if priceInfo.BreakQuantity == 100 {
					partToAdd.Cost_100pc.Float64 = priceInfo.UnitPrice
					partToAdd.Cost_100pc.Valid = true
				} else if priceInfo.BreakQuantity == 1000 {
					partToAdd.Cost_1000pc.Float64 = priceInfo.UnitPrice
					partToAdd.Cost_1000pc.Valid = true
				}
			}
			if !partToAdd.Cost_1pc.Valid {
				partToAdd.Cost_1pc.Float64 = 0
				partToAdd.Cost_1pc.Valid = true
			}
			if !partToAdd.Cost_100pc.Valid {
				partToAdd.Cost_100pc.Float64 = 0
				partToAdd.Cost_100pc.Valid = true
			}
			if !partToAdd.Cost_1000pc.Valid {
				partToAdd.Cost_1000pc.Float64 = 0
				partToAdd.Cost_1000pc.Valid = true
			}
			updateTable(partToAdd, db) //Update the table
		}
		fmt.Println("Sleeping for 2 seconds to not hit the query limit") //Reusing the Mouser snooze function
		time.Sleep(2 * time.Second)                                      //snooze
	}
	db.Close()
}

// // DIgikey add new part
func newPart_Digikey(partToFind newPart) {
	db := dbConnection()
	genTokens()
	var partToAdd updatePart
	err := godotenv.Load("../FrontEnd/.env") //using .env file for easier access
	if err != nil {
		log.Fatal("Error loading .env file") //error handling
	}
	err = godotenv.Load("../FrontEnd/tokens.env") //using .env file for easier access
	if err != nil {
		log.Fatal("Error loading tokens.env file") //error handling
	}

	//Retreive the part number
	partToAdd.Supplier_Part_Number.String = partToFind.Supplier_Part_Number
	partToAdd.Supplier_Part_Number.Valid = true
	digiKeyPartNumber := partToFind.Supplier_Part_Number
	partSearchOptions := "string"

	//Access with api Key
	clientID := os.Getenv("DIGI_ID")
	accessToken := "Bearer " + os.Getenv("ACCESS_TOKEN")

	partURL := "https://api.digikey.com/Search/v3/Products/" + partToAdd.Supplier_Part_Number.String

	// //For testing purposes
	// fmt.Println(clientID)
	// fmt.Println(accessToken)
	// //Print URL for testing Purposes, TODO: remove or comment with integration
	// fmt.Println("URL:>", partURL)

	// create a map to hold the request data
	requestData := map[string]interface{}{
		"SearchByPartRequest": map[string]string{
			"digiKeyPartNumber": digiKeyPartNumber,
			"partSearchOptions": partSearchOptions,
		},
	}

	// convert the map to a JSON byte array
	jsonStr, err := json.Marshal(requestData)
	req, err := http.NewRequest("GET", partURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", accessToken)
	req.Header.Set("X-DIGIKEY-Client-Id", clientID)
	//reponse for post requestion
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	//Response should always be 200
	fmt.Println("response Status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	var results1 digiKeyResponse
	///Unmarshal results to more readable data
	err = json.Unmarshal(body, &results1)
	if err != nil {
		log.Fatal(err)
	}

	// pretty print json on terminal for testing
	// data, err := json.MarshalIndent(results1, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(data))

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Received HTTP status: %s, likely could not find the part.\n", resp.Status)
	} else {
		fmt.Println("Response is 200 OK")
		partToAdd.Description.String = results1.DetailedDescription
		partToAdd.Description.Valid = true
		partToAdd.Manufacturer.String = results1.Manufacturer.Value
		partToAdd.Manufacturer.Valid = true
		partToAdd.Manufacturer_Part_Number.String = results1.ManufacturerPartNumber
		partToAdd.Manufacturer_Part_Number.Valid = true
		partToAdd.Supplier_Part_Number.String = results1.DigiKeyPartNumber
		partToAdd.Part_Category.String = results1.Category.Value
		partToAdd.Part_Category.Valid = true
		if results1.RoHSStatus == "ROHS3 Compliant" {
			partToAdd.RoHS.Bool = true
		} else {
			partToAdd.RoHS.Bool = false
		}
		partToAdd.RoHS.Valid = true
		partToAdd.PrimaryVendorStock.Int64 = int64(results1.QuantityAvailable)
		partToAdd.PrimaryVendorStock.Valid = true
		for _, priceInfo := range results1.StandardPricing {
			if priceInfo.BreakQuantity == 1 {
				partToAdd.Cost_1pc.Float64 = priceInfo.UnitPrice
				partToAdd.Cost_1pc.Valid = true
			} else if priceInfo.BreakQuantity == 100 {
				partToAdd.Cost_100pc.Float64 = priceInfo.UnitPrice
				partToAdd.Cost_100pc.Valid = true
			} else if priceInfo.BreakQuantity == 1000 {
				partToAdd.Cost_1000pc.Float64 = priceInfo.UnitPrice
				partToAdd.Cost_1000pc.Valid = true
			}
		}
		if !partToAdd.Cost_1pc.Valid {
			if results1.UnitPrice != 0 {
				partToAdd.Cost_1pc.Float64 = results1.UnitPrice
				partToAdd.Cost_1pc.Valid = true
			} else {
				partToAdd.Cost_1pc.Float64 = 0
				partToAdd.Cost_1pc.Valid = true
			}
		}
		if !partToAdd.Cost_100pc.Valid {
			partToAdd.Cost_100pc.Float64 = 0
			partToAdd.Cost_100pc.Valid = true
		}
		if !partToAdd.Cost_1000pc.Valid {
			partToAdd.Cost_1000pc.Float64 = 0
			partToAdd.Cost_1000pc.Valid = true
		}
		insertTable(partToAdd, "Digi-Key", db) //Update the table
	}
	fmt.Println("Sleeping for 2 seconds to not hit the query limit") //Reusing the Mouser snooze function
	time.Sleep(2 * time.Second)                                      //snooze
	// fmt.Println(partToAdd)
	db.Close()
}

// Update mouser list
func update_Mouser(partsList []int) {
	db := dbConnection()
	// partsList := (gatherDatabase(db))
	for i := range partsList { //i in range parts list, if only one part to update it will only do it once
		var partToAdd updatePart
		partToAdd.ID = partsList[i]
		err := godotenv.Load("../FrontEnd/.env") //using .env file for easier access
		if err != nil {
			log.Fatal("Error loading .env file") //error handling
		}
		//Get Mouser API key
		APIKey := os.Getenv("MOUSER_API")
		//Access with api Key
		partURL := "https://api.mouser.com/api/v1/search/partnumber?apiKey=" + APIKey
		//Print URL for testing Purposes,
		// fmt.Println("URL:>", partURL)
		var supplierPartNumber string
		//get the old information from the list
		err = db.QueryRow("SELECT `Supplier Part Number 1` FROM `lusher engineering parts database`.electronics_parts WHERE `ID`=?", partsList[i]).Scan(&supplierPartNumber)
		if err != nil {
			fmt.Println("Could not query Database")
			log.Fatal(err)
		}
		//Retreive the part number
		partToAdd.Supplier_Part_Number.String = supplierPartNumber
		partToAdd.Supplier_Part_Number.Valid = true
		mouserPartNumber := supplierPartNumber
		partSearchOptions := "string"

		// create a map to hold the request data
		requestData := map[string]interface{}{
			"SearchByPartRequest": map[string]string{
				"mouserPartNumber":  mouserPartNumber,
				"partSearchOptions": partSearchOptions,
			},
		}

		// convert the map to a JSON byte array
		jsonStr, err := json.Marshal(requestData)
		req, err := http.NewRequest("POST", partURL, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		//reponse for post requestion
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		//Response should always be 200
		fmt.Println("response Status:", resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		var results1 mouserResponse
		///Unmarshal results to more readable data
		err = json.Unmarshal(body, &results1)
		if err != nil {
			log.Fatal(err)
		}

		// // pretty print json on terminal for testing
		// data, err := json.MarshalIndent(results1, "", "  ")
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println(string(data))
		// fmt.Println(results1.SearchResults.NumberOfResult)

		//TODO integration: remove all commented print commands used for testing
		//If no results
		if results1.SearchResults.NumberOfResult == 0 {
			fmt.Println("No Results Found For Supplier Part Number: ", mouserPartNumber)
			continue
		} else if results1.SearchResults.NumberOfResult == 1 { //exactly one result
			for _, rec := range results1.SearchResults.Parts {
				// fmt.Println("Description: ", rec.Description)
				partToAdd.Description.String = rec.Description
				partToAdd.Description.Valid = true
				// fmt.Println("Manufacturer: ", rec.Manufacturer)
				partToAdd.Manufacturer.String = string(rec.Manufacturer)
				partToAdd.Manufacturer.Valid = true
				// fmt.Println("ManufacturerPartNumber: ", rec.ManufacturerPartNumber)
				partToAdd.Manufacturer_Part_Number.String = rec.ManufacturerPartNumber
				partToAdd.Manufacturer_Part_Number.Valid = true
				// fmt.Println("MouserPartNumber: ", rec.MouserPartNumber)
				partToAdd.Supplier_Part_Number.String = rec.MouserPartNumber
				partToAdd.Supplier_Part_Number.Valid = true
				// fmt.Println("Category: ", rec.Category)
				partToAdd.Part_Category.String = rec.Category
				partToAdd.Part_Category.Valid = true
				// fmt.Println("ROHSStatus: ", rec.ROHSStatus)
				if rec.ROHSStatus == "RoHS Compliant" {
					partToAdd.RoHS.Bool = true
				} else {
					partToAdd.RoHS.Bool = false
				}
				partToAdd.RoHS.Valid = true
				// fmt.Println(rec.Availability)
				// fmt.Println(rec.AvailabilityInStock)
				// fmt.Println(rec.AvailabilityOnOrder)
				// fmt.Println("FactoryStock: ", rec.AvailabilityInStock)
				if rec.AvailabilityInStock != "" {
					intVar, err := strconv.ParseInt(rec.AvailabilityInStock, 10, 64)
					if err != nil {
						log.Fatal(err)
					}
					partToAdd.PrimaryVendorStock.Int64 = intVar
					partToAdd.PrimaryVendorStock.Valid = true
				} else {
					partToAdd.PrimaryVendorStock.Int64 = 0
					partToAdd.PrimaryVendorStock.Valid = true
				}
				//Parse this array of pricebreaks
				for _, priceBreak := range rec.PriceBreaks {

					if priceBreak.Quantity == 1 {
						priceStr := priceBreak.Price
						priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
						priceFloat, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							log.Fatal(err)
						}
						partToAdd.Cost_1pc.Float64 = priceFloat
						partToAdd.Cost_1pc.Valid = true
					} else if priceBreak.Quantity == 100 {
						priceStr := priceBreak.Price
						priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
						priceFloat, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							log.Fatal(err)
						}
						partToAdd.Cost_100pc.Float64 = priceFloat
						partToAdd.Cost_100pc.Valid = true
					} else if priceBreak.Quantity == 1000 {
						priceStr := priceBreak.Price
						priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
						priceFloat, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							log.Fatal(err)
						}
						partToAdd.Cost_1000pc.Float64 = priceFloat
						partToAdd.Cost_1000pc.Valid = true
					}
				}
				if !partToAdd.Cost_1pc.Valid {
					partToAdd.Cost_1pc.Float64 = 0
					partToAdd.Cost_1pc.Valid = true
				}
				if !partToAdd.Cost_100pc.Valid {
					partToAdd.Cost_100pc.Float64 = 0
					partToAdd.Cost_100pc.Valid = true
				}
				if !partToAdd.Cost_1000pc.Valid {
					partToAdd.Cost_1000pc.Float64 = 0
					partToAdd.Cost_1000pc.Valid = true
				}
				// fmt.Println("DataSheetURL", rec.DataSheetURL) //ask if he wants this
			}
		} else { //More than one result
			for _, rec := range results1.SearchResults.Parts {
				// fmt.Println(string(rec.MouserPartNumber))
				if supplierPartNumber == string(rec.MouserPartNumber) {
					// fmt.Println("Description: ", rec.Description)
					partToAdd.Description.String = rec.Description
					partToAdd.Description.Valid = true
					// fmt.Println("Manufacturer: ", rec.Manufacturer)
					partToAdd.Manufacturer.String = string(rec.Manufacturer)
					partToAdd.Manufacturer.Valid = true
					// fmt.Println("ManufacturerPartNumber: ", rec.ManufacturerPartNumber)
					partToAdd.Manufacturer_Part_Number.String = rec.ManufacturerPartNumber
					partToAdd.Manufacturer_Part_Number.Valid = true
					// fmt.Println("MouserPartNumber: ", rec.MouserPartNumber)
					partToAdd.Supplier_Part_Number.String = rec.MouserPartNumber
					partToAdd.Supplier_Part_Number.Valid = true
					// fmt.Println("Category: ", rec.Category)
					partToAdd.Part_Category.String = rec.Category
					partToAdd.Part_Category.Valid = true
					// fmt.Println("ROHSStatus: ", rec.ROHSStatus)
					if rec.ROHSStatus == "RoHS Compliant" {
						partToAdd.RoHS.Bool = true
					} else {
						partToAdd.RoHS.Bool = false
					}
					partToAdd.RoHS.Valid = true
					partToAdd.Part_Category.String = rec.Category
					partToAdd.Description.Valid = true
					// fmt.Println(rec.Availability)
					// fmt.Println(rec.AvailabilityInStock)
					// fmt.Println(rec.AvailabilityOnOrder)
					// fmt.Println("FactoryStock: ", rec.AvailabilityInStock)
					if rec.AvailabilityInStock != "" {
						intVar, err := strconv.ParseInt(rec.AvailabilityInStock, 10, 64)
						if err != nil {
							log.Fatal(err)
						}
						partToAdd.PrimaryVendorStock.Int64 = intVar
						partToAdd.PrimaryVendorStock.Valid = true
					} else {
						partToAdd.PrimaryVendorStock.Int64 = 0
						partToAdd.PrimaryVendorStock.Valid = true
					}
					//Parse this array of pricebreaks
					for _, priceBreak := range rec.PriceBreaks {
						// fmt.Println(priceBreak.Quantity)
						if priceBreak.Quantity == 1 {
							priceStr := priceBreak.Price
							priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
							priceFloat, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
							}
							partToAdd.Cost_1pc.Float64 = priceFloat
							partToAdd.Cost_1pc.Valid = true
						} else if priceBreak.Quantity == 100 {
							priceStr := priceBreak.Price
							priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
							priceFloat, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
							}
							partToAdd.Cost_100pc.Float64 = priceFloat
							partToAdd.Cost_100pc.Valid = true
						} else if priceBreak.Quantity == 1000 {
							priceStr := priceBreak.Price
							priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
							priceFloat, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								log.Fatal(err)
							}
							partToAdd.Cost_1000pc.Float64 = priceFloat
							partToAdd.Cost_1000pc.Valid = true
						}
					}
					if !partToAdd.Cost_1pc.Valid {
						partToAdd.Cost_1pc.Float64 = 0
						partToAdd.Cost_1pc.Valid = true
					}
					if !partToAdd.Cost_100pc.Valid {
						partToAdd.Cost_100pc.Float64 = 0
						partToAdd.Cost_100pc.Valid = true
					}
					if !partToAdd.Cost_1000pc.Valid {
						partToAdd.Cost_1000pc.Float64 = 0
						partToAdd.Cost_1000pc.Valid = true
					}
				}
				// fmt.Println("DataSheetURL", rec.DataSheetURL) //ask if he wants this
				//Update the table
			}
			// }
			updateTable(partToAdd, db)
		}
		fmt.Println("Sleeping for 2 seconds to not hit the query limit") //mouser allows 30 results a minute
		time.Sleep(2 * time.Second)                                      //snooze
	}
	db.Close()
}

// function to add a new part
func newPart_Mouser(partToFind newPart) {
	db := dbConnection()
	var partToAdd updatePart
	//Load up environment variables
	err := godotenv.Load("../FrontEnd/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	APIKey := os.Getenv("MOUSER_API")
	//Access with api Key
	partURL := "https://api.mouser.com/api/v1/search/partnumber?apiKey=" + APIKey
	// data := mouserParams{
	// 	mouserPartNumber:  partNumber,
	// 	partSearchOptions: "string",
	// }
	fmt.Println("URL:>", partURL)
	// var jsonStr = []byte(`{  "SearchByPartRequest": {    "mouserPartNumber": "653-G4W-1114P-DC12",    "partSearchOptions": "string"  }}`)

	partToAdd.Supplier_Part_Number.String = partToFind.Supplier_Part_Number
	partToAdd.Supplier_Part_Number.Valid = true
	mouserPartNumber := partToFind.Supplier_Part_Number
	partSearchOptions := "string"

	// create a map to hold the request data
	requestData := map[string]interface{}{
		"SearchByPartRequest": map[string]string{
			"mouserPartNumber":  mouserPartNumber,
			"partSearchOptions": partSearchOptions,
		},
	}

	// convert the map to a JSON byte array
	jsonStr, err := json.Marshal(requestData)
	req, err := http.NewRequest("POST", partURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	var results1 mouserResponse
	err = json.Unmarshal(body, &results1)
	if err != nil {
		log.Fatal(err)
	}

	// // pretty print json on terminal
	// data, err := json.MarshalIndent(results1, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(data))
	// fmt.Println(results1.SearchResults.NumberOfResult)

	//NO results
	if results1.SearchResults.NumberOfResult == 0 {
		fmt.Println("No Results Found For Supplier Part Number: ", mouserPartNumber)
	} else if results1.SearchResults.NumberOfResult == 1 { //One result
		for _, rec := range results1.SearchResults.Parts {
			// fmt.Println("Description: ", rec.Description)
			partToAdd.Description.String = rec.Description
			partToAdd.Description.Valid = true
			// fmt.Println("Manufacturer: ", rec.Manufacturer)
			partToAdd.Manufacturer.String = string(rec.Manufacturer)
			partToAdd.Manufacturer.Valid = true
			// fmt.Println("ManufacturerPartNumber: ", rec.ManufacturerPartNumber)
			partToAdd.Manufacturer_Part_Number.String = rec.ManufacturerPartNumber
			partToAdd.Manufacturer_Part_Number.Valid = true
			// fmt.Println("MouserPartNumber: ", rec.MouserPartNumber)
			partToAdd.Supplier_Part_Number.String = rec.MouserPartNumber
			partToAdd.Supplier_Part_Number.Valid = true
			// fmt.Println("Category: ", rec.Category)
			partToAdd.Part_Category.String = rec.Category
			partToAdd.Part_Category.Valid = true
			// fmt.Println("ROHSStatus: ", rec.ROHSStatus)
			if rec.ROHSStatus == "RoHS Compliant" {
				partToAdd.RoHS.Bool = true
			} else {
				partToAdd.RoHS.Bool = false
			}
			partToAdd.RoHS.Valid = true
			partToAdd.Part_Category.String = rec.Category
			partToAdd.Description.Valid = true
			// fmt.Println(rec.Availability)
			// fmt.Println(rec.AvailabilityInStock)
			// fmt.Println(rec.AvailabilityOnOrder)
			// fmt.Println("FactoryStock: ", rec.AvailabilityInStock)
			if rec.AvailabilityInStock != "" {
				intVar, err := strconv.ParseInt(rec.AvailabilityInStock, 10, 64)
				if err != nil {
					log.Fatal(err)
				}
				partToAdd.PrimaryVendorStock.Int64 = intVar
				partToAdd.PrimaryVendorStock.Valid = true
			} else {
				partToAdd.PrimaryVendorStock.Int64 = 0
				partToAdd.PrimaryVendorStock.Valid = true
			}
			//Parse this array of pricebreaks
			for _, priceBreak := range rec.PriceBreaks {

				if priceBreak.Quantity == 1 {
					priceStr := priceBreak.Price
					priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
					priceFloat, err := strconv.ParseFloat(priceStr, 64)
					if err != nil {
						log.Fatal(err)
					}
					partToAdd.Cost_1pc.Float64 = priceFloat
					partToAdd.Cost_1pc.Valid = true
				} else if priceBreak.Quantity == 100 {
					priceStr := priceBreak.Price
					priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
					priceFloat, err := strconv.ParseFloat(priceStr, 64)
					if err != nil {
						log.Fatal(err)
					}
					partToAdd.Cost_100pc.Float64 = priceFloat
					partToAdd.Cost_100pc.Valid = true
				} else if priceBreak.Quantity == 1000 {
					priceStr := priceBreak.Price
					priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
					priceFloat, err := strconv.ParseFloat(priceStr, 64)
					if err != nil {
						log.Fatal(err)
					}
					partToAdd.Cost_1000pc.Float64 = priceFloat
					partToAdd.Cost_1000pc.Valid = true
				}
			}
			if !partToAdd.Cost_1pc.Valid {
				partToAdd.Cost_1pc.Float64 = 0
				partToAdd.Cost_1pc.Valid = true
			}
			if !partToAdd.Cost_100pc.Valid {
				partToAdd.Cost_100pc.Float64 = 0
				partToAdd.Cost_100pc.Valid = true
			}
			if !partToAdd.Cost_1000pc.Valid {
				partToAdd.Cost_1000pc.Float64 = 0
				partToAdd.Cost_1000pc.Valid = true
			}
			// fmt.Println("DataSheetURL", rec.DataSheetURL) //ask if he wants this
			insertTable(partToAdd, "Mouser Electronics", db) // call insert table instead of update
		}
	} else { //More than one result
		// for _, searchResult := range results1.SearchResults {
		for _, rec := range results1.SearchResults.Parts {
			// fmt.Println(string(rec.MouserPartNumber))
			if partToAdd.Manufacturer_Part_Number.String == string(rec.MouserPartNumber) {
				// fmt.Println("Description: ", rec.Description)
				partToAdd.Description.String = rec.Description
				partToAdd.Description.Valid = true
				// fmt.Println("Manufacturer: ", rec.Manufacturer)
				partToAdd.Manufacturer.String = string(rec.Manufacturer)
				partToAdd.Manufacturer.Valid = true
				// fmt.Println("ManufacturerPartNumber: ", rec.ManufacturerPartNumber)
				partToAdd.Manufacturer_Part_Number.String = rec.ManufacturerPartNumber
				partToAdd.Manufacturer_Part_Number.Valid = true
				// fmt.Println("MouserPartNumber: ", rec.MouserPartNumber)
				partToAdd.Supplier_Part_Number.String = rec.MouserPartNumber
				partToAdd.Supplier_Part_Number.Valid = true
				// fmt.Println("Category: ", rec.Category)
				partToAdd.Part_Category.String = rec.Category
				partToAdd.Part_Category.Valid = true
				// fmt.Println("ROHSStatus: ", rec.ROHSStatus)
				if rec.ROHSStatus == "RoHS Compliant" {
					partToAdd.RoHS.Bool = true
				} else {
					partToAdd.RoHS.Bool = false
				}
				partToAdd.RoHS.Valid = true
				partToAdd.Part_Category.String = rec.Category
				partToAdd.Description.Valid = true
				// fmt.Println(rec.Availability)
				// fmt.Println(rec.AvailabilityInStock)
				// fmt.Println(rec.AvailabilityOnOrder)
				// fmt.Println("FactoryStock: ", rec.AvailabilityInStock)
				if rec.AvailabilityInStock != "" {
					intVar, err := strconv.ParseInt(rec.AvailabilityInStock, 10, 64)
					if err != nil {
						log.Fatal(err)
					}
					partToAdd.PrimaryVendorStock.Int64 = intVar
					partToAdd.PrimaryVendorStock.Valid = true
				} else {
					partToAdd.PrimaryVendorStock.Int64 = 0
					partToAdd.PrimaryVendorStock.Valid = true
				}
				//Parse this array of pricebreaks
				for _, priceBreak := range rec.PriceBreaks {
					// fmt.Println(priceBreak.Quantity)
					if priceBreak.Quantity == 1 {
						priceStr := priceBreak.Price
						priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
						priceFloat, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							log.Fatal(err)
						}
						partToAdd.Cost_1pc.Float64 = priceFloat
						partToAdd.Cost_1pc.Valid = true
					} else if priceBreak.Quantity == 100 {
						priceStr := priceBreak.Price
						priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
						priceFloat, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							log.Fatal(err)
						}
						partToAdd.Cost_100pc.Float64 = priceFloat
						partToAdd.Cost_100pc.Valid = true
					} else if priceBreak.Quantity == 1000 {
						priceStr := priceBreak.Price
						priceStr = strings.ReplaceAll(priceStr, "$", "") // remove "$" character
						priceFloat, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							log.Fatal(err)
						}
						partToAdd.Cost_1000pc.Float64 = priceFloat
						partToAdd.Cost_1000pc.Valid = true
					}
				}
				if !partToAdd.Cost_1pc.Valid {
					partToAdd.Cost_1pc.Float64 = 0
					partToAdd.Cost_1pc.Valid = true
				}
				if !partToAdd.Cost_100pc.Valid {
					partToAdd.Cost_100pc.Float64 = 0
					partToAdd.Cost_100pc.Valid = true
				}
				if !partToAdd.Cost_1000pc.Valid {
					partToAdd.Cost_1000pc.Float64 = 0
					partToAdd.Cost_1000pc.Valid = true
				}
			}
			// fmt.Println("DataSheetURL", rec.DataSheetURL) //ask if he wants this
		}
		insertTable(partToAdd, "Mouser Electronics", db) // call insert table instead of update
	}
	fmt.Println("Sleeping for 2 seconds to not hit the query limit") //mouser allows 30 results a minute
	time.Sleep(2 * time.Second)
	db.Close()
}

// Update the table if already exists
func updateTable(partToAdd updatePart, db *sql.DB) {
	// var internalPN int
	// err := db.QueryRow("SELECT MAX(`Internal PN`) FROM `lusher engineering parts database`.electronics_parts;").Scan(&internalPN)
	// if err != nil {
	// 	log.Fatal(err) // Handle error as appropriate for your application
	// }
	// internalPN++

	result, err := db.Exec("update `electronics_parts` set `Part Description` = ?, Manufacturer = ?, `Manufacturer Part Number` = ?, `Supplier Part Number 1` = ?, `Primary Vendor Stock` = ?, `RoHS Compliant` = ?,  `Part Category` = ?,`Cost 1pc` = ?, `Cost 100pc` = ?, `Cost 1000pc` = ?  where id = ?", partToAdd.Description.String, partToAdd.Manufacturer.String, partToAdd.Manufacturer_Part_Number.String, partToAdd.Supplier_Part_Number.String, partToAdd.PrimaryVendorStock.Int64, partToAdd.RoHS.Bool, partToAdd.Part_Category.String, partToAdd.Cost_1pc, partToAdd.Cost_100pc, partToAdd.Cost_1000pc, partToAdd.ID)
	if err != nil {
		//
		log.Fatal(err)
	} else {
		if result != nil {
			log.Println("Database Updated")
		}
	}
}

// Insert new row into table
func insertTable(partToAdd updatePart, vendor string, db *sql.DB) {
	//Get the last used ID so we can increment it
	row := db.QueryRow("SELECT id FROM `lusher engineering parts database`.electronics_parts ORDER BY id DESC LIMIT 1;")
	var id int
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			// handle no rows found
		} else {
			log.Fatal(err)
		}
	}
	newID := id + 1
	var internalPN int
	err := db.QueryRow("SELECT MAX(`Internal PN`) FROM `lusher engineering parts database`.electronics_parts;").Scan(&internalPN)
	if err != nil {
		log.Fatal(err) // Handle error as appropriate for your application
	}
	internalPN++
	query := "INSERT INTO `lusher engineering parts database`.electronics_parts (`id`, `Internal PN`, `Part Description`, Manufacturer, `Manufacturer Part Number`,`Supplier 1`, `Supplier Part Number 1`, `Primary Vendor Stock`, `RoHS Compliant`, `Part Verified`, `Part Category`, `Cost 1pc`, `Cost 100pc`, `Cost 1000pc`, `Auto Update`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	fmt.Println()
	_, err = db.ExecContext(context.Background(), query, newID, internalPN, partToAdd.Description.String, partToAdd.Manufacturer.String, partToAdd.Manufacturer_Part_Number.String, vendor, partToAdd.Supplier_Part_Number.String, partToAdd.PrimaryVendorStock.Int64, partToAdd.RoHS.Bool, 0, partToAdd.Part_Category.String, partToAdd.Cost_1pc, partToAdd.Cost_100pc, partToAdd.Cost_1000pc, 0)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal("Could not insert as ID", err)
	}
	log.Printf("Inserted part as id: %d", newID)
}

// Authorize through OAUTH for digi-key
func genTokens() {
	err := godotenv.Load("../FrontEnd/.env") //using .env file for easier access
	if err != nil {
		log.Fatal("Error loading .env file") //error handling
	}
	err = godotenv.Load("../FrontEnd/tokens.env") //using .env file for easier access
	if err != nil {
		log.Fatal("Error loading tokens.env file") //error handling
	}
	//First, check if valid refresh token:
	// create a map to hold the request data
	requestData := map[string]string{
		"client_id":     os.Getenv("DIGI_ID"),
		"client_secret": os.Getenv("DIGI_SECRET"),
		"refresh_token": os.Getenv("REFRESH_TOKEN"),
		"grant_type":    "refresh_token",
	}

	// Convert the map to a URL-encoded string
	data := url.Values{}
	for key, value := range requestData {
		data.Set(key, value)
	}

	// URL to which you want to make the POST request
	url := "https://api.digikey.com/v1/oauth2/token"
	// Create the HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the Content-Type header to specify that the request body is in "application/x-www-form-urlencoded" format
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the HTTP request and get the response
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 401 {
			getWebCode()
		} else {
			fmt.Println("Request failed with status code:", resp.StatusCode)
		}
	} else {
		// Read the response body
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		// Use the response data as needed
		fmt.Println("Response:", response)

		//Store the new tokens
		access_token := response["access_token"].(string)
		refresh_token := response["refresh_token"].(string)

		// Update the .env file with the new values
		updateEnv("ACCESS_TOKEN", access_token)
		updateEnv("REFRESH_TOKEN", refresh_token)
	}
}

func getWebCode() {
	// Define the redirect URI where the response will be sent
	redirectURI := "https://localhost" // Update with your actual redirect URI

	// Create a channel to signal when the code has been retrieved
	codeChan := make(chan string)

	// Create a web server to listen for the response
	http.HandleFunc("/code", func(w http.ResponseWriter, r *http.Request) {
		// Parse the query string
		query := r.URL.Query()
		code := query.Get("code")
		errorMessage := query.Get("error")

		if errorMessage != "" {
			// Handle the error response
			fmt.Println("Error response:", errorMessage)
			codeChan <- "" // Signal that an error occurred
		} else if code != "" {
			// Handle the authorization code response
			fmt.Println("Authorization code:", code)
			codeChan <- code
		}

		// Respond to the browser
		fmt.Fprintf(w, "Response received.")
	})

	// Start the web server on the specified redirect URI
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println("Error starting web server:", err)
		}
	}()

	// Construct the OAuth2 authorization URL
	oauthURL := "https://api.digikey.com/v1/oauth2/authorize?response_type=code&client_id=" + os.Getenv("DIGI_ID") + "&redirect_uri=" + redirectURI

	// Open the URL in the default web browser
	err := open.Run(oauthURL)
	if err != nil {
		// Handle error if the URL couldn't be opened
		fmt.Println("Error opening the URL.")
	}
	// Wait for the code to be retrieved or an error to occur
	select {
	case code := <-codeChan:
		if code == "" {
			// If an error occurred, allow for user input as a backup
			fmt.Println("Unable to retrieve the code from the browser. Please enter the code manually from:\n", oauthURL)
			var userCode string
			fmt.Scan(&userCode)
			code = userCode
		}
		// Call the function to exchange the code for tokens
		exchangeCodeForTokens(code, oauthURL)
	case <-time.After(10 * time.Second): // Timeout after 10 seconds
		fmt.Println("Timed out waiting for the authorization code.")
		fmt.Println("Unable to retrieve the code from the browser. Please enter the code manually from:\n", oauthURL)
		var userCode string
		fmt.Scan(&userCode)
		exchangeCodeForTokens(userCode, oauthURL)
	}

}

func exchangeCodeForTokens(digiCode string, oauthURL string) {
	err := godotenv.Load("../FrontEnd/.env") //using .env file for easier access
	if err != nil {
		log.Fatal("Error loading .env file") //error handling
	}
	requestData := map[string]string{
		"code":          digiCode,
		"client_id":     os.Getenv("DIGI_ID"),
		"client_secret": os.Getenv("DIGI_SECRET"),
		"redirect_uri":  "https://localhost",
		"grant_type":    "authorization_code",
	}
	// Convert the map to a URL-encoded string
	data := url.Values{}
	for key, value := range requestData {
		data.Set(key, value)
	}
	// URL to which you want to make the POST request
	url := "https://api.digikey.com/v1/oauth2/token"
	// Create the HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	// Set the Content-Type header to specify that the request body is in "application/x-www-form-urlencoded" format
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the HTTP request and get the response
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// fmt.Println("Request failed with status code:", resp.StatusCode)
		fmt.Println("Something went wrong, please enter code again (only valid for 60 seconds):\n", oauthURL)
		var userCode string
		exchangeCodeForTokens(userCode, oauthURL)
		fmt.Scan(&userCode)
		return
	}
	// Read the response body
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	// Use the response data as needed
	fmt.Println("Response:", response)

	//Store the new tokens
	access_token := response["access_token"].(string)
	refresh_token := response["refresh_token"].(string)

	// Update the .env file with the new values
	updateEnv("ACCESS_TOKEN", access_token)
	updateEnv("REFRESH_TOKEN", refresh_token)
}

func updateEnv(key, value string) {
	// Read the .env file content
	envFile, err := ioutil.ReadFile("../FrontEnd/tokens.env")
	if err != nil {
		fmt.Println("Error reading .env file:", err)
		return
	}

	lines := strings.Split(string(envFile), "\n")
	updated := false

	// Iterate through the lines and replace the one with the given key
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = key + "=" + value
			updated = true
			break
		}
	}

	// If the key doesn't exist, append it to the end
	if !updated {
		lines = append(lines, key+"="+value)
	}

	// Convert the lines back to a single string
	newEnvFile := strings.Join(lines, "\n")

	// Write the updated content back to the .env file
	err = ioutil.WriteFile("../FrontEnd/tokens.env", []byte(newEnvFile), 0644)
	if err != nil {
		fmt.Println("Error writing .env file:", err)
		return
	}

	os.Setenv(key, value)

}

//export AddNewPartCSharpInput
func AddNewPartCSharpInput(newPartNumber *C.char, supplierName *C.char) {
	goPartNumber := C.GoString(newPartNumber)
	goSupplierName := C.GoString(supplierName)
	addNewPart(goSupplierName, goPartNumber)
}

//export updatePartArray
func updatePartArray(arr *C.int, len C.int, supplierName *C.char) {
	// Convert C array to Go slice
	goLen := int(len) // convert C.int to int
	goSlice := make([]int, goLen)
	for i := 0; i < goLen; i++ {
		goSlice[i] = int((*(*[1 << 28]C.int)(unsafe.Pointer(arr)))[i])
	}

	goSupplierName := C.GoString(supplierName)

	if goSupplierName == "Digi-Key" {
		update_Digikey(goSlice)
	} else {
		update_Mouser(goSlice)
	}
}

// Timer for timing purposes
func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}

// Main
// Left test cases and examples in per client request
func main() {
	// defer timer("main")()
	// fmt.Println("Database opened successfully!")
	// digiKey_Test := []int{7}
	// update_Digikey(digiKey_Test)
	// testPart_digi := "74AHC1G14SE-7DITR-ND"
	// addNewPart("Digi-Key", testPart_digi)
	//test
	// partsList := (gatherDatabase(db))
	// updateAll(partsList, "digikey")
	// updateMany_Digikey(partsList)
	// testPart := "863-NTH4L022N120M3S"
	// testPart := "151-203-RC"

	//ID 524
	// fmt.Println(partsList)
	//Add New Part  863-NTH4L022N120M3S
	// testPart := "G4W-1114P-US-TV8-HP-DC12"
	// testPart := "AS321KTR-G1"
	// addNewPart("Mouser Electronics", testPart)
	// mouserDemo := []int{38, 87, 143, 144, 169, 175, 182, 295, 331, 334}
	// update_Mouser(mouserDemo, db)

}
