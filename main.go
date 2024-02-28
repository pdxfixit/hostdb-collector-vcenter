package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	// load config
	loadConfig()
	if Config.Debug {
		log.Println(fmt.Sprintf("%v", Config))
	}

	// get the inventory
	inventoryBytes, err := httpRequest("GET", Config.Inventory.Url, nil, nil)
	if err != nil {
		log.Println(fmt.Sprintf("%s", inventoryBytes))
		log.Fatal(err)
	}

	// unmarshal into our struct
	inventory := Inventory{}
	if err := json.Unmarshal(inventoryBytes, &inventory); err != nil {
		log.Fatal(err)
	}
	log.Println(fmt.Sprintf("Found %d instances. Not all of these are vcenters.", len(inventory.Urls)))
	if Config.Debug {
		log.Println(fmt.Sprintf("%v", inventory.Urls))
	}

	// loop over the vcenters in the inventory
	for n, item := range inventory.Urls {
		if item.Type.Name != "vcenter" {
			continue
		}

		// output to file; skip if already created
		if _, err := os.Stat(fmt.Sprintf("/Users/pdxfixit/issues/vcenter/output/%s/_list.json", item.Url)); os.IsExist(err) {
			continue
		}

		// INSECURE
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		//
		// login to a vcenter, get a session token
		//

		// encode creds for basic auth
		plaintextVcenterCreds := fmt.Sprintf("%s:%s", Config.Vcenter.User, Config.Vcenter.Pass)
		encodedVcenterCreds := base64.StdEncoding.EncodeToString([]byte(plaintextVcenterCreds))

		// get a session token
		log.Println(fmt.Sprintf("Trying %s (%d/%d)...", item.Url, n, len(inventory.Urls)))
		session, err := httpRequest(
			"POST",
			fmt.Sprintf("https://%s/rest/com/vmware/cis/session", item.Url),
			nil,
			map[string]string{
				"Authorization": fmt.Sprintf("Basic %s", encodedVcenterCreds),
			},
		)
		if err != nil {
			log.Println(fmt.Sprintf("%s", session))
			log.Println(err)
			continue
		}

		// unmarshal the response into a struct
		sessionToken := VcenterSessionToken{}
		if err := json.Unmarshal(session, &sessionToken); err != nil {
			log.Println(err)
			continue
		}
		vmwareSessionHeader := map[string]string{"vmware-api-session-id": sessionToken.Value}
		if Config.Debug {
			log.Println(fmt.Sprintf("%v", vmwareSessionHeader))
		}

		//
		// collect data from this vcenter
		//

		hostdbRecords := HostdbDocumentSet{
			"vcenter-vm",
			time.Now().UTC().Format("2006-01-02 15:04:05"),
			map[string]interface{}{
				"vcenter-url":      item.Url,
				"vcenter-desc":     item.Desc,
				"vcenter-name":     item.Name,
				"vcenter-location": item.Location,
			},
			"hostdb-collector-vcenter",
			[]HostdbDocument{},
		}

		// get the list of VMs
		log.Println("Getting a list of VMs...")
		listBytes, err := httpRequest("GET", fmt.Sprintf("https://%s/rest/vcenter/vm/", item.Url), nil, vmwareSessionHeader)
		if err != nil {
			log.Println(fmt.Sprintf("%s", listBytes))
			log.Println(err)
			continue
		}

		// unmarshal it into the struct
		vmList := VcenterList{}
		if err := json.Unmarshal(listBytes, &vmList); err != nil {
			log.Println(err)
			continue
		}
		log.Println(fmt.Sprintf("Retrieved a list with %d VMs.", len(vmList.Value)))
		if Config.Debug {
			log.Println(fmt.Sprintf("%v", vmList.Value))

			// output
			if err := os.MkdirAll(fmt.Sprintf("/Users/pdxfixit/issues/vcenter/output/%s", item.Url), 0755); err != nil {
				log.Fatal(err)
			}
			if err := ioutil.WriteFile(
				fmt.Sprintf("/Users/pdxfixit/issues/vcenter/output/%s/_list.json", item.Url),
				listBytes,
				0644,
			); err != nil {
				log.Fatal(err)
			}
		}

		// get the data for each of the VMs in the list
		for i, vm := range vmList.Value {

			if vm.Id == "" {
				continue
			}

			url := fmt.Sprintf("https://%s/rest/vcenter/vm/%s", item.Url, vm.Id)

			log.Println(fmt.Sprintf("Fetching data for VM '%s' (%d/%d)...", vm.Id, i, len(vmList.Value)))
			vmBytes, err := httpRequest(
				"GET",
				url,
				nil,
				vmwareSessionHeader,
			)
			if err != nil {
				log.Println(fmt.Sprintf("%s", vmBytes))
				log.Println(err)
				continue
			}

			// unmarshal the response into a struct
			vmData := VcenterVm{}
			if err := json.Unmarshal(vmBytes, &vmData); err != nil {
				log.Println(err)
				continue
			}

			// re-marshal only the value property
			json, err := json.Marshal(vmData.Value)
			if err != nil {
				log.Println(err)
				continue
			}

			if Config.Debug {
				log.Println(fmt.Sprintf("%s", json))

				// output
				if err := os.MkdirAll(fmt.Sprintf("/Users/pdxfixit/issues/vcenter/output/%s", item.Url), 0755); err != nil {
					log.Fatal(err)
				}
				if err := ioutil.WriteFile(
					fmt.Sprintf("/Users/pdxfixit/issues/vcenter/output/%s/%s.json", item.Url, vm.Id),
					vmBytes,
					0644,
				); err != nil {
					log.Fatal(err)
				}
			}

			// TODO: validate data

			doc := HostdbDocument{
				"",
				"",
				"",
				"",
				time.Now().UTC().Format("2006-01-02 15:04:05"),
				"",
				nil,
				json,
				"",
			}

			hostdbRecords.Records = append(hostdbRecords.Records, doc)
		}

		//
		// logout from vcenter, destroy session
		//

		log.Println("Closing vCenter session...")
		if deleteResponse, err := httpRequest(
			"DELETE",
			fmt.Sprintf("https://%s/rest/com/vmware/cis/session", item.Url),
			nil,
			vmwareSessionHeader,
		); err != nil {
			log.Println(fmt.Sprintf("%s", deleteResponse))
			log.Println(err)
			continue
		}

		//
		// post data to HostDB
		//

		// marshal the filled struct into bytes for transmission
		requestBytes, err := json.Marshal(hostdbRecords)
		if err != nil {
			log.Println(err)
			continue
		}

		// encode creds for basic auth
		plaintextHostdbCreds := fmt.Sprintf("%s:%s", Config.Hostdb.User, Config.Hostdb.Pass)
		encodedHostdbCreds := base64.StdEncoding.EncodeToString([]byte(plaintextHostdbCreds))

		// do the needful
		log.Println(fmt.Sprintf("Posting data as '%s' to HostDB @ %s/records/...", Config.Hostdb.User, Config.Hostdb.Url))
		hostdbResponse, err := httpRequest(
			"POST",
			fmt.Sprintf("%s/records/", Config.Hostdb.Url),
			bytes.NewReader(requestBytes),
			map[string]string{
				"Authorization": fmt.Sprintf("Basic %s", encodedHostdbCreds),
			},
		)
		if err != nil {
			log.Println(fmt.Sprintf("%s", hostdbResponse))
			log.Println(err)
			continue
		}
		log.Println("Data saved to HostDB.")

		// TODO: analyze the response from HostDB.
		//  Even if we do nothing for now, we should log anything other than "ok".

	}

	log.Println("All done!")

}

func httpRequest(method string, url string, body io.Reader, header map[string]string) (bytes []byte, err error) {

	var res *http.Response

	// Client
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatal(err)
	}

	// INSECURE
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// headers
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}

	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		return bytes, errors.New(res.Status)
	}

	return bytes, nil

}
