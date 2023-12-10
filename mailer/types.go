package mailer

type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption   string
	FromAddress string
	FromName    string
	Jobs        chan Message
	Results     chan Result
	API         string
	APIKey      string
	APIUrl      string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template   string
	Attachments []string
	Data        interface{}
}

type Result struct {
	Success bool
	Error   error
}
