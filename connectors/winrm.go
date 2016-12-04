package connectors

import (
	"bufio"
	"bytes"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/projectjane/jane/models"
	"github.com/projectjane/jane/parse"
	"github.com/projectjane/winrm"
)

//WinRM Struct representing a WinRM Connector
type WinRM struct {
}

//Listen Listen not implemented for WinRM
func (x WinRM) Listen(commandMsgs chan<- models.Message, connector models.Connector) {
	defer Recovery(connector)
	for _, command := range connector.Commands {
		if command.RunCheck {
			if connector.Debug {
				log.Print("Starting Listener for " + connector.ID + " " + command.Name)
			}
			go checkRM(commandMsgs, command, connector)
		}
	}
}

//Command Standard command parser
func (x WinRM) Command(message models.Message, publishMsgs chan<- models.Message, connector models.Connector) {
	for _, command := range connector.Commands {
		if match, tokens := parse.Match(command.Match, message.In.Text); match {
			args := parse.Substitute(command.Args, tokens)

			tokens["STDOUT"] = sendCommand(command.Cmd, args, connector)

			var color = "NONE"
			var match = false
			if match, _ = parse.Match(command.Green, tokens["STDOUT"]); match {
				color = "SUCCESS"
			}
			if match, _ = parse.Match(command.Yellow, tokens["STDOUT"]); match {
				color = "WARN"
			}
			if match, _ = parse.Match(command.Red, tokens["STDOUT"]); match {
				color = "FAIL"
			}
			message.In.Tags = parse.TagAppend(message.In.Tags, connector.Tags+","+command.Tags)
			message.Out.Text = connector.ID + " " + command.Name
			message.Out.Detail = parse.Substitute(command.Output, tokens)
			message.Out.Status = color
			publishMsgs <- message
		}
	}
}

//Publish Not implemented for WinRM
func (x WinRM) Publish(publishMsgs <-chan models.Message, connector models.Connector) {
	return
}

//Help Returns help information for the connector
func (x WinRM) Help(connector models.Connector) (help []string) {
	help = make([]string, 0)
	for _, command := range connector.Commands {
		if !command.HideHelp {
			if command.Help != "" {
				help = append(help, command.Help)
			} else {
				help = append(help, strings.Replace(command.Match, "*", "", -1))
			}
		}
	}
	return help
}

func checkRM(commandMsgs chan<- models.Message, command models.Command, connector models.Connector) {

	// command vars
	var state = command.Green
	var stateReset = true
	var counter = 1
	var interval = 1
	var sampling = 1
	var remind = 0
	if command.Interval > 0 {
		interval = command.Interval
	}
	if command.Sampling > 0 {
		sampling = command.Sampling
	}
	if command.Remind > 0 {
		remind = command.Remind
	}

	// loop commands
	for {

		// reset vars
		var color = "NONE"
		var match = false
		var newstate = ""
		var sendAlert = false

		// make the call
		out := sendCommand(command.Cmd, command.Args, connector)

		// interpret results
		if match, _ = parse.Match(command.Green, out); match {
			newstate = command.Green
			color = "SUCCESS"
		}
		if match, _ = parse.Match(command.Yellow, out); match {
			newstate = command.Yellow
			color = "WARN"
		}
		if match, _ = parse.Match(command.Red, out); match {
			newstate = command.Red
			color = "FAIL"
		}

		// handle state change
		if newstate != state {

			if stateReset {
				counter = 1
				stateReset = false
			}

			// sampling
			if counter == sampling {
				sendAlert = true
			}

			// change to green
			if newstate == command.Green {
				sendAlert = true
			}

		}

		// handle non-green state
		if newstate != command.Green && counter == remind && remind > 1 {
			sendAlert = true

		}

		// send message
		if sendAlert {
			var tokens = parse.Tokens()
			tokens["STDOUT"] = out
			var message models.Message
			message.In.ConnectorType = connector.Type
			message.In.ConnectorID = connector.ID
			message.In.Tags = parse.TagAppend(connector.Tags, command.Tags)
			message.In.Process = false
			message.Out.Text = connector.ID + " " + command.Name
			message.Out.Detail = parse.Substitute(command.Output, tokens)
			message.Out.Status = color
			commandMsgs <- message
			state = newstate
			counter = 0
			stateReset = true
		}

		// wait
		counter += 1
		time.Sleep(time.Duration(interval) * time.Minute)

	}

}

func sendCommand(command, args string, connector models.Connector) string {
	port := 5985
	if connector.Port != "" {
		port, _ = strconv.Atoi(connector.Port)
	}

	endpoint := winrm.NewEndpoint(connector.Server, port, false, false, nil, nil, nil, 0)
	rmclient, err := winrm.NewClient(endpoint, connector.Login, connector.Pass)
	if err != nil {
		log.Println("Error connecting to endpoint:", err)
		return "Error connecting to endpoint: " + err.Error()
	}

	var in bytes.Buffer
	var out bytes.Buffer
	var e bytes.Buffer

	stdin := bufio.NewReader(&in)
	stdout := bufio.NewWriter(&out)
	stderr := bufio.NewWriter(&e)

	_, err = rmclient.RunWithInput(command, stdout, stderr, stdin)
	if err != nil {
		return "Error running command: " + err.Error()
	}

	if e.String() != "" {
		return e.String()
	}

	return out.String()
}
