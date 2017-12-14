package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

// ConfigsModel ...
type ConfigsModel struct {
	//oAuth
	token  string
	roomID string

	//onSuccess
	fromName string
	message  string
	color    string
	notify   string

	//onFail
	fromNameOnError string
	messageOnError  string
	colorOnError    string
	notifyOnError   string

	//settings
	messageFormat string
	baseURL       string

	isBuildFailedMode string
}

// -----------------------
// --- Functions
// -----------------------

func createConfigsModelFromEnvs() ConfigsModel {
	return ConfigsModel{
		token:  os.Getenv("auth_token"),
		roomID: os.Getenv("room_id"),

		fromName: os.Getenv("from_name"),
		message:  os.Getenv("message"),
		color:    os.Getenv("color"),
		notify:   os.Getenv("notify"),

		fromNameOnError: os.Getenv("from_name_on_error"),
		messageOnError:  os.Getenv("message_on_error"),
		colorOnError:    os.Getenv("color_on_error"),
		notifyOnError:   os.Getenv("notify_on_error"),

		messageFormat: os.Getenv("message_format"),
		baseURL:       os.Getenv("server_url"),

		isBuildFailedMode: os.Getenv("BITRISE_BUILD_STATUS"),
	}
}

func (configs ConfigsModel) print() {
	log.Infof("Configs:")

	log.Printf("- token: %s", "***")
	log.Printf("- roomID: %s", configs.roomID)

	log.Printf("- fromName: %s", configs.fromName)
	log.Printf("- message: %s", configs.message)
	log.Printf("- color: %s", configs.color)
	log.Printf("- notify: %s", configs.notify)

	log.Printf("- fromNameOnError: %s", configs.fromNameOnError)
	log.Printf("- messageOnError: %s", configs.messageOnError)
	log.Printf("- colorOnError: %s", configs.colorOnError)
	log.Printf("- notifyOnError: %s", configs.notifyOnError)

	log.Printf("- messageFormat: %s", configs.messageFormat)
	log.Printf("- baseURL: %s", configs.baseURL)
}

// -----------------------
// --- Main
// -----------------------

func main() {

	fmt.Println()

	config := createConfigsModelFromEnvs()

	config.print()

	isFailed := (config.isBuildFailedMode != "0")

	if isFailed {
		config.fromName = config.fromNameOnError
		config.message = config.messageOnError
		config.color = config.colorOnError
		config.notify = config.notifyOnError
	}

	//
	// Create request
	fmt.Println()
	log.Infof("Performing request")
	fmt.Println()

	values := url.Values{
		"room_id":        {config.roomID},
		"from":           {config.fromName},
		"message":        {config.message},
		"color":          {config.color},
		"message_format": {config.messageFormat},
		"notify":         {config.notify},
	}

	valuesReader := *strings.NewReader(values.Encode())

	url := config.baseURL + "/room/" + config.roomID + "/notification?auth_token=" + config.token

	request, err := http.NewRequest("POST", url, &valuesReader)

	if err != nil {
		log.Errorf("failed to perform request, error: %s", err)
		os.Exit(1)
	}

	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		// On error, any Response can be ignored
		log.Errorf("failed to perform request, error: %s", err)
		os.Exit(1)
	}

	// The client must close the response body when finished with it
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Errorf("Failed to close response body, error: %s", err)
			os.Exit(1)
		}
	}()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Errorf("failed to read response body, error: %s", err)
		os.Exit(1)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode > http.StatusMultipleChoices {
		log.Errorf("non success status code (%d)", response.StatusCode)
		log.Printf("body: %s", body)
		os.Exit(1)
	}

	// Success
	log.Donef("Request successful")

	fmt.Println()

	log.Infof("Response:")
	log.Printf("status code: %d", response.StatusCode)
	log.Printf("body: %s", body)

	fmt.Println()

}
