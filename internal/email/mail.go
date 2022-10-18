// Package for describing email structure thus handling emails
// with the general porpuse of generating a json file from
// a directory of email files
package email

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Struct to describe header for all the records/emails
type EmailsDir struct {
	Index   string   `json:"index"`
	Records []*Email `json:"records"`
}

// defines Email file as an object to store its information as json format
type Email struct {
	Origin      string   `json:"origin"`
	SubFolder   string   `json:"sub_folder"`
	Id          string   `json:"id"`
	Date        string   `json:"date"`
	From        string   `json:"from"`
	To          []string `json:"to"`
	Cc          []string `json:"cc"`
	Bcc         []string `json:"bcc"`
	Subject     string   `json:"subject"`
	Version     string   `json:"version"`
	ContentType string   `json:"type"`
	Encoding    string   `json:"encoding"`
	Content     string   `json:"content"`
}

// defines scan states when reading email file line by line
var States = [10]string{
	"Message-ID: ",
	"Date: ",
	"From: ",
	"To: ",
	"Cc: ",
	"Bcc: ",
	"Subject: ",
	"Mime-Version: ",
	"Content-Type: ",
	"Content-Transfer-Encoding: ",
}

// by a given email file line, returns the scan state if found, otherwise, returns the provided previus state
func GetState(line, prev_state string) string {

	for i := 0; i < len(States); i++ {

		if strings.HasPrefix(line, States[i]) {
			return States[i]
		}
	}

	return prev_state

}

// by a given string of email adresses, removes spaces, tabs, new lines and returns an array
// where each element is an email address
func GetMails(mails_string string) []string {

	// temp_line :=

	return strings.FieldsFunc(
		strings.TrimFunc(
			strings.Replace(
				strings.Replace(mails_string,
					" ", "", -1), "\t", "", -1), func(c rune) bool {
				return c == 9 || c == 10 || c == 11
			}),
		func(c rune) bool {
			return c == ','
		})

}

// by a given directory path of email files, processes each one and
// creates a json file
func DirToJson(path string) {
	var emails []*Email

	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		if !d.IsDir() {

			email := PathToEmail(path)
			emails = append(emails, &email)
		}
		return nil
	})

	p, _ := json.Marshal(EmailsDir{Index: "emails", Records: emails})
	_ = ioutil.WriteFile("mailsdir.json", p, 0644)

}

// by a given email file path, returns an Email object
func PathToEmail(path string) Email {

	readFile, err := os.Open(path)
	if err != nil {
		fmt.Printf("couldn't open email file at path: %s\n", path)
		panic(err)

	}

	var ID string
	var date string
	var from_mail string
	var to_mails string
	var cc_mails string
	var subject string
	var mime_version string
	var content_type string
	var encoding string
	var bcc_mails string

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	prev_state := ""

	for fileScanner.Scan() {

		line := fileScanner.Text()

		state := GetState(line, prev_state)

		if !strings.HasPrefix(line, "X-") {

			switch state {

			case "Message-ID: ":
				ID += line

			case "Date: ":
				date += line

			case "From: ":
				from_mail += line

			case "To: ":
				to_mails += line

			case "Cc: ":
				cc_mails += line

			case "Subject: ":
				subject += line

			case "Mime-Version: ":
				mime_version += line

			case "Content-Type: ":
				content_type += line

			case "Content-Transfer-Encoding: ":
				encoding += line

			case "Bcc: ":
				bcc_mails += line

			default:

			}

		} else {
			break
		}

		prev_state = state

	}

	readFile.Close()

	splited_path := strings.Split(path, "\\")

	sub_folder := strings.Join(splited_path[3:len(splited_path)-1], `\`)
	raw_content, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Printf("couldn't read file at path: %s\n", path)
		panic(err)

	}

	//to get the message content of the email file, separates
	//the whole content by a regular expression

	reg_delimeter := regexp.MustCompile(`X-FileName\s*(.*?)\s*\n`)
	raw_content_str := string(raw_content)
	content := reg_delimeter.Split(raw_content_str, -1)

	return Email{
		Origin:      splited_path[2],
		SubFolder:   sub_folder,
		Id:          strings.TrimPrefix(ID, "Message-ID: "),
		Date:        strings.TrimPrefix(date, "Date: "),
		From:        strings.TrimPrefix(from_mail, "From: "),
		To:          GetMails(strings.TrimPrefix(to_mails, "To: ")),
		Cc:          GetMails(strings.TrimPrefix(cc_mails, "Cc: ")),
		Bcc:         GetMails(strings.TrimPrefix(bcc_mails, "Bcc: ")),
		Subject:     strings.TrimPrefix(subject, "Subject: "),
		Version:     strings.TrimPrefix(mime_version, "Mime-Version: "),
		ContentType: strings.TrimPrefix(content_type, "Content-Type: "),
		Encoding:    strings.TrimPrefix(encoding, "Content-Transfer-Encoding: "),
		Content:     content[1],
	}
}