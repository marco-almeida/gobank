package service

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/marco-almeida/mybank/internal/config"
// 	"github.com/stretchr/testify/require"
// )

// func TestSendEmailWithGmail(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip()
// 	}

// 	config, err := config.LoadConfig("../..")
// 	require.NoError(t, err)

// 	fmt.Println(config)

// 	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
// 	fmt.Println(sender.(*GmailSender))

// 	subject := "A test email"
// 	content := `
// 	<h1>Hello world</h1>
// 	`
// 	to := []string{"marco.aa.almeida02@gmail.com"}

// 	err = sender.SendEmail(subject, content, to, nil, nil, nil)
// 	require.NoError(t, err)
// }
