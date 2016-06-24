package connectors

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/projectjane/jane/models"
)

type Jira struct {
}

func (x Jira) Listen(commandMsgs chan<- models.Message, connector models.Connector) {
	return
}

func (x Jira) Command(message models.Message, publishMsgs chan<- models.Message, connector models.Connector) {
	if strings.HasPrefix(strings.ToLower(message.In.Text), strings.ToLower("jira create")) {
		createJiraIssue(message, publishMsgs, connector)
	} else {
		parseJiraIssue(message, publishMsgs, connector)
	}
}

func (x Jira) Publish(connector models.Connector, message models.Message, target string) {
	return
}

func (x Jira) Help(connector models.Connector) (help string) {
	help += "jira - mention a jira ticket and it'll be displayed\n"
	help += "jira create {project key} {summary}"
	return help
}

type ticket struct {
	Key    string `json:"key"`
	Fields fields `json:"fields"`
}

type fields struct {
	Summary     string   `json:"summary"`
	Status      status   `json:"status"`
	Description string   `json:"description"`
	Priority    priority `json:"priority"`
	Assignee    assignee `json:"assignee"`
}

type status struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

type priority struct {
	Name string `json:"name"`
}

type assignee struct {
	DisplayName string `json:"displayName"`
}

type createObject struct {
	Fields createFields `json:"fields"`
}

type createFields struct {
	Project   project   `json:"project"`
	Summary   string    `json:"summary"`
	IssueType issueType `json:"issueType"`
}

type project struct {
	Key string `json:"key"`
}

type issueType struct {
	Name string `json:"name"`
}

type createdIssue struct {
	Id   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

func createJiraIssue(message models.Message, publishMsgs chan<- models.Message, connector models.Connector) {
	msg := strings.TrimSpace(strings.Replace(message.In.Text, "jira create", "", 1))
	fields := strings.Fields(msg)
	summary := strings.Join(fields[2:], " ")
	client := &http.Client{}
	auth := encodeB64(connector.Login + ":" + connector.Pass)

	issuetype := issueType{
		Name: fields[0],
	}

	project := project{
		Key: fields[1],
	}

	issueFields := createFields{
		Project:   project,
		Summary:   summary,
		IssueType: issuetype,
	}

	issue := createObject{
		Fields: issueFields,
	}

	issueJson, err := json.Marshal(issue)
	if err != nil {
		log.Println(err)

		return
	}

	req, err := http.NewRequest("POST", "https://"+connector.Server+"/rest/api/2/issue", bytes.NewBuffer(issueJson))
	if err != nil {
		log.Print(err)
		message.Out.Text = "Failed to create issue"
		publishMsgs <- message
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+auth)

	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Print(err)
	}

	var created createdIssue
	json.Unmarshal(body, &created)

	message.Out.Text = created.Key
	publishMsgs <- message
}

func parseJiraIssue(message models.Message, publishMsgs chan<- models.Message, connector models.Connector) {
	var jiraRegex = regexp.MustCompile("[a-zA-Z]{2,12}-[0-9]{1,10}")
	issues := jiraRegex.FindAllString(message.In.Text, -1)
	for _, issue := range issues {
		if connector.Debug {
			log.Print("Jira match: " + issue)
		}

		client := &http.Client{}
		auth := encodeB64(connector.Login + ":" + connector.Pass)
		req, err := http.NewRequest("GET", "https://"+connector.Server+"/rest/api/2/issue/"+issue, nil)
		if err != nil {
			log.Print(err)
			return
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Basic "+auth)

		response, err := client.Do(req)
		if err != nil {
			log.Print(err)
			return
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Print(err)
		}
		var ticket ticket
		json.Unmarshal(body, &ticket)
		if connector.Debug {
			log.Printf("Jira result: %+v", ticket)
		}
		if ticket.Fields.Status.Name == "" {
			return
		}
		message.Out.Link = "https://" + connector.Server + "/browse/" + issue
		message.Out.Text = strings.ToUpper(issue) + " - " + ticket.Fields.Summary
		message.Out.Detail = fmt.Sprintf("Status: %s\nPriority: %s\nAssignee: %s\n",
			ticket.Fields.Status.Name, ticket.Fields.Priority.Name, ticket.Fields.Assignee.DisplayName)
		publishMsgs <- message
	}
}

func encodeB64(message string) string {
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(base64Text, []byte(message))
	return string(base64Text)
}
