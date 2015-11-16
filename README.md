# Jane

Jane is a bot to pull information and conduct operational activities in your chatops scenario - even in a command line way. This bot is written in go and is made to be configuration driven. Contributions are welcome via pull requests. If you want to know why the name 'Jane' was chosen, talk to @kcwinner.



## Getting Started
* This is developed using Go 1.5.1
* Pull the project with 'go get github.com/mmcquillan/jane'
* Compile with 'go install jane.go'
* Use the sample.config for an example configuration file
* Use the startup.conf as an upstart script to start/stop/restart on Linux systems
* You can rename your bot by setting the top-level Name configuration setting


## Installation
* Copy the compiled jane to /usr/bin/jane
* Copy the startup.conf file to /etc/init/jane.conf
* Make a configuration file and locate it at /etc/jane.config.json
* Startup as "sudo service start jane"


## Configuration
The entire configuration of the site is done via a json config file. The configuration file is expected to be named 'jane.config' and will be looked for in this order:
* -config someconfig.json - Pass in a configuration file location as a command line parameter
* ./jane.config - the location of the jane binary
* ~/jane.config - the home directory of the user
* /etc/jane.config - the global config


## Connectors
Connectors are what Jane uses to pull in information, interpret them and issue out a response. The Routes specify where the results from the input should be written to or * for all. The Target can specify a channel in the case of slack. To add a new connector, Put them in the connectors folder and make an entry in connectors/list.go.

For the connector configuration, when adding routes, you must specify the ID of the connector you want to route response to.

### Command Line Connector
`{"Type": "cli", "ID": "cli", "Active": false,
 "Routes": [{"Match": "*", "Connectors": "cli", "Target": ""}]}`

### Slack Connector
Note that the slack connector has a default route that will always route traffic from the source within slack back to itself, such that it can respond per channel and DM's. This cannot be turned off, but caution as DM's could be routed to places they should not, so best to leave this without routes.
`{"Type": "slack", "ID": "slack", "Active": true,
    "Key": "<SlackToken>", "Image": ":game_die:"
  }`

### RSS Connector
`{"Type": "rss", "ID": "Bamboo Build", "Active": true,
    "Server": "https://BambooUser:BambooPass@somecompany.atlassian.net/builds/plugins/servlet/streams?local=true",
    "SuccessMatch": "successful", "FailureMatch": "fail",
    "Routes": [
      {"Match": "*", "Connectors": "slack", "Target": "#devops"},
      {"Match": "NextGen", "Connectors": "slack", "Target": "#nextgen"}
    ]
  }`

 `{"Type": "rss", "ID": "AWS EC2", "Active": true,
    "Server": "http://status.aws.amazon.com/rss/ec2-us-east-1.rss",
    "Routes": [
      {"Match": "*", "Connectors": "slack", "Target": "#devops"}
    ]
  }`

### Email Connector
`{"Type": "email", "ID": "Email Server", "Active": true,
    "Server": "mail.someserver.com", "Login": "jane", "Pass": "abc123",
    "From": "jane@janecorp.com"
 }`

### Monitor Connector
Note, this is currently setup to execute a nagios style monitoring script and interpret the results as the example shows below.

`{"Type": "monitor", "ID": "Elasticsearch Node", "Active": true,
    "Server": "elasticsearch1.somecompany.com", "Login": "jane", "Pass": "abc123",
    "SuccessMatch": "OK", "WarningMatch": "WARNING", "FailureMatch": "CRITICAL",
    "Checks": [
      {"Name": "Apt Check", "Check": "/usr/lib/nagios/plugins/check_apt"},
      {"Name": "Disk Check", "Check": "/usr/lib/nagios/plugins/check_disk -w10% -c5% -A"},
      {"Name": "Elasticsearch Check", "Check": "/usr/lib/nagios/plugins/check_procs -a elasticsearch -c1:1"}
    ],
    "Routes": [
      {"Match": "*", "Connectors": "slack", "Target": "#devops"},
      {"Match": "*", "Connectors": "slack", "Target": "@matt"}
    ]
  }`


### Webserver Connector
This will check if a given website returns a 200 status, else it will alert.

`{"Type": "website", "ID": "Website Monitor", "Active": true,
      "Checks": [
        {"Name": "Google", "Check": "https://google.com"},
        {"Name": "Yahoo", "Check": "http://www.yahoo.com"}
      ],
      "Routes": [
        {"Match": "*", "Connectors": "*"}
      ]
    }`


## Commands
Commands are what execute or respond to requests by listeners.

### Response Command
`{"Type": "response", "Match": "motivate", "Output": "You can _do it_ %msg%!"}`

### Exec Command
`{"Type": "exec", "Match": "big", "Output": "```%stdout%```", "Cmd": "/usr/bin/figlet", "Args": "-w80 %msg%"}`

### Help Command
`{"Type": "help", "Match": "help"}`

### Reload Command
`{"Type": "reload", "Match": "reload", "Output": "Reloading command configuration"}`


