package main

import (
	"fmt"
	"log"
	"net/http"
	"crypto/sha512"
	"encoding/base64"
	"strings"
	"strconv"
	"bytes"
	"sync"
	"time"
	"context"
)

/* TODO: Dhitaj
1. Add API protection in future.
2. Unit Test Cases
3. Handle Shutdown function properly
4. Currently logging is maitained at same level, later we will make it to different levels based on the need.
*/

/**********************************************************************************
** 
** Following section contains the global variables.
**
**********************************************************************************/

// Following variable is used to return the sequence number on hash request.
// Later this is used in retriving the actual hash value of the password.
var hashCounter uint64

// Following map maintains the map of hash counter values and hash values of passwords.
// Customer calls the hash REST API with sequence number and we return the hash value of password from this map.
// We are not persisting the valus of hashes on disk based on the requirements given.
var hashMapData map[uint64]string
var hashCounterWaitGroup sync.WaitGroup

var performanceMetric uint64
var performanceMetricWaitGroup sync.WaitGroup

/**********************************************************************************/

/**
 * HomePage
 *
 * This function does nothing except returning a note to customer.
 *
 * TODO: We can return a list of REST APIs exposed by this service to customer.
 */
func HomePage(response http.ResponseWriter, request *http.Request) {
	log.Print("Processing Home Page Request.")

	fmt.Fprintf(response, "Password Hash REST APIs in Golang")
}

/**
 * InsertPasswordInHashMap
 *
 * This function evaluates the SHA512 hash value of given password and inserts
 * the password hash value in hashMapData corresponding to given sequence value.
 *
 * password: Attribute containing plain password.
 * hashCounterValue: This is the sequence number generated when customer called the hash REST API.
 *
 * Note: This is private function and is called asynchronously by hash REST API.
 */
func InsertPasswordInHashMap(password string, hashCounterValue uint64) {
	log.Print("Inside InsertPasswordInHashMap.")

	// SHA512 hash function was given as part of requirements.
	hashFunction := sha512.New()
	hashFunction.Write([]byte(password))

	// Converting the hashed value to Base64 format.
	sha512_hash := base64.URLEncoding.EncodeToString(hashFunction.Sum(nil))

	// Now storing the value in HASH map for later usage.
	hashMapData[hashCounterValue] = sha512_hash

	log.Print("Exiting InsertPasswordInHashMap.")
}

/**
 * Hash --> /hash REST API
 *
 * This API helps customer to retrive the HASH value of given password.
 *
 * response: HTTP Response
 * request: HTTP Request
 *
 * TODO: Add following validations.
 * 1. Maximum length check of password. (Currently I'm not sure about the requirement here. Ww will reject the packet based on the length in the future.)
 * 2. Currently I'm assuming that data always come in given format i.e. password=<Password>, later we will add the validation based on all possible combinations.
 *
 */
func Hash(response http.ResponseWriter, request *http.Request) {
	startTime := time.Now()

	log.Print("Processing Hash Request.")

	// Getting the body of REST API to retrive the password.
	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	requestBodyStr := buf.String()

	// TODO Dhiraj: Validate the password format.
	// Retriving the password.
	password := strings.TrimPrefix(requestBodyStr, "password")
	
	// Synchronizing the counter increment operation.
	// Get the next hash counter value so that we can send back that value to customer.
	hashCounterWaitGroup.Add(1)
	hashCounter = hashCounter + 1
	hashCounterValue := hashCounter
	hashCounterWaitGroup.Done()
	hashCounterWaitGroup.Wait()

	// Return the hash counter value to customer.
	str := fmt.Sprint(hashCounterValue)
	fmt.Fprintf(response, str)
	
	// Add the password hash value asynchronously. 
	go InsertPasswordInHashMap(password, hashCounterValue)

	endTime := time.Now()

	go AddPerformanceNumber(endTime.Sub(startTime))

}

/**
 * GetHashData --> /hash/{id} REST API
 *
 * This API helps customer to retrive the HASH value of given password using ID returned in HASH REST API.
 *
 * response: HTTP Response
 * request: HTTP Request
 *
 * TODO: Add following validations.
 * 1. Maximum length check of password. (Currently I'm not sure about the requirement here. Ww will reject the packet based on the length in the future.)
 * 2. Currently I'm assuming that data always come in given format i.e. password=<Password>, later we will add the validation based on all possible combinations.
 *
 */
func GetHashData(response http.ResponseWriter, request *http.Request) {
	startTime := time.Now()

	log.Print("Processing Get Hash Data Request.")


	// Get the hash counter value from the URL.
	idStr := strings.TrimPrefix(request.URL.Path, "/hash/")
	id, err := strconv.ParseUint(idStr, 10, 64)

	// If it is not a valid number, return an error to caller.
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)

		log.Print("Invalid Hash ID")
		fmt.Fprintf(response, "Error: Invalid Hash ID")
	} else {
		if hashData, found := hashMapData[id]; found {
			log.Print("Returning Hash Data")
			fmt.Fprintf(response, hashData)
		}  else {
			// If caller passes unknown hash counter number, return an error to caller.
			response.WriteHeader(http.StatusBadRequest)

			log.Print("Invalid Hash ID")
			fmt.Fprintf(response, "Error: Invalid Hash ID")
		}

	}

	endTime := time.Now()

	go AddPerformanceNumber(endTime.Sub(startTime))
}

/**
 * AddPerformanceNumber
 *
 * This function adds the performance numbers of the APIs.
 * In this version we are evaluating the performance of just Hash REST APIs.
 *
 * timeTaken: Time taken by REST API to process the request.
 *
 */
 func AddPerformanceNumber(timeTaken time.Duration) {
	log.Print("Inside AddPerformanceNumber.")

	performanceMetricWaitGroup.Add(1)

	performanceMetric += uint64(timeTaken)

	performanceMetricWaitGroup.Done()
	performanceMetricWaitGroup.Wait()


	log.Print("Exiting AddPerformanceNumber.")
}

/**
 *
 * Stats --> /stats REST API
 *
 * This API returns the performance metric of REST APIs.
 *
 */
func Stats(response http.ResponseWriter, request *http.Request) {
	log.Print("Processing stats Request.")

	var metric uint64

	metric = 0
	if (performanceMetric > 0) {
		metric = performanceMetric/hashCounter;
	}
	
	metricStr := fmt.Sprint(metric)
	hashCounterStr := fmt.Sprint(hashCounter)

	text := `{"total": ` + hashCounterStr + `, "average": ` + metricStr + `}`
	log.Print(text)
	fmt.Fprintf(response, text)
}


/**
 *
 * HandleRequests
 *
 * This function initializes the REST APIs and initiating the REST API server at port 7777.
 *
 * TODO Dhiraj: Take port number from the command line argument.
 */
func HandleRequests() {
	handler := http.NewServeMux()
    server := http.Server{Addr: ":7777", Handler: handler}
    handler.HandleFunc("/shutdown", func(response http.ResponseWriter, request *http.Request) {
		response.Write([]byte("OK"))
		
        go func() {
            if err := server.Shutdown(context.Background()); err != nil {
                log.Fatal(err)
            }
        }()
	})
	
	handler.HandleFunc("/", HomePage)
	handler.HandleFunc("/hash", Hash)
	handler.HandleFunc("/hash/", GetHashData)
	handler.HandleFunc("/stats", Stats)

	log.Fatal(server.ListenAndServe())
}

/**
 * main
 *
 * Entry point of service. Initializing the REST APIs end points and global variables.
 *
 */
func main() {
	hashCounter = 0
	hashMapData = make(map[uint64]string)

	HandleRequests()
}
