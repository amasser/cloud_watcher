package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var limit int64 = 50
var totalGroups int
var mygroups []string
var retention string
var zeroRetention []string
var messages []string
var departmentTag string
var environmentTag string
var provisionerTag string

type ListTagsLogGroupOutput struct {

	// The tags for the log group.
	Tags map[string]*string `locationName:"tags" min:"1" type:"map"`
	// contains filtered or unexported fields
}

func main() {

	b, err := ioutil.ReadFile("creds.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1hPtvEQngzZ1TvMB7cqGEf-5sJUivdb81OFERtPTOWfg"

	writeRange := "A3"

	var vr sheets.ValueRange

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := cloudwatchlogs.New(sess)
	cmdGroups(svc)

	for i := range zeroRetention {
		myval := []interface{}{zeroRetention[i]}
		vr.Values = append(vr.Values, myval)

	}

	// fmt.Println(zeroRetention)

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("RAW").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}

	// fmt.Println(zeroRetention)
	// fmt.Println((len(zeroRetention)))
	fmt.Printf(strconv.Itoa(totalGroups) + " log groups here... \n")

	fmt.Printf(strconv.Itoa((len(zeroRetention))) + " log groups have unlimited retention... \n")

	var vr2 sheets.ValueRange
	writeRange = "E1"
	messages = append(messages, (strconv.Itoa(totalGroups) + " log groups here... \n" + (strconv.Itoa((len(zeroRetention))) + " log groups have unlimited retention.")))
	// fmt.Println(messages)
	myMessage := []interface{}{messages[0]}
	vr2.Values = append(vr2.Values, myMessage)

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr2).ValueInputOption("RAW").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
}

func cmdGroups(client *cloudwatchlogs.CloudWatchLogs) error {
	req := cloudwatchlogs.DescribeLogGroupsInput{
		Limit: aws.Int64(limit),
	}

	handler := func(res *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		totalGroups = totalGroups + (len(res.LogGroups))
		for _, group := range res.LogGroups {
			name := group.LogGroupName
			// logGroupName := fmt.Sprintf(`%s`, *name)
			tags := cloudwatchlogs.ListTagsLogGroupInput{
				LogGroupName: name,
			}
			tagsOutput, err := client.ListTagsLogGroup(&tags)
			if err != nil {
				//fmt.Println(tagsOutput)
			}

			if group.RetentionInDays != nil {
				retention = strconv.FormatInt(*group.RetentionInDays, 10)
				switch retentionDays := *group.RetentionInDays; {
				case int(retentionDays) > 90:
					fmt.Println("over 90!")
				}
			} else {
				retention = "0"
				department := (*tagsOutput).Tags["Department"]
				if department != nil {
					departmentTag = *department
				} else {
					departmentTag = ""
				}
				provisioner := (*tagsOutput).Tags["Provisioner"]
				if provisioner != nil {
					provisionerTag = *provisioner
				} else {
					provisionerTag = ""
				}
				nameTag := *name
				// sheetsString := fmt.Sprintf("q%, s%", *name, departmentTag)
				sheetsString := (nameTag + ":" + departmentTag + ":" + provisionerTag)
				fmt.Println(sheetsString)
				stringerBell := fmt.Sprintf(sheetsString)
				zeroRetention = append(zeroRetention, stringerBell)
			}
			mygroups = append(mygroups, *name, retention)

		}
		return true // want more pages
	}

	err := client.DescribeLogGroupsPages(&req, handler)
	if err != nil {
		return err
	}

	// fmt.Println(mygroups)
	// fmt.Println(zeroRetention)
	return nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
