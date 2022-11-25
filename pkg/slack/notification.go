package slack

import (
        "mongoex/cmd/config"
        "fmt"
	"github.com/slack-go/slack"
        "encoding/json"
        "strconv"
        "time"
)

func Notification(message string) error {
    _, _, slackWebhookUrl := config.ParseConfig()
    // if slackwebhook is not defined, then exit
    if slackWebhookUrl == "" {
            return nil
    }

    // slack attachment (can customise)
    attachment := slack.Attachment{
	Fallback:      message,
	AuthorName:    "mongobot",
	Text:          message,
	Footer:        "mongoex",
	FooterIcon:    "https://platform.slack-edge.com/img/default_application_icon.png",
	Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
    }

    // the message from attachment (above)
    msg := slack.WebhookMessage{
	Attachments: []slack.Attachment{attachment},
    }

    // send to slack
    err := slack.PostWebhook(slackWebhookUrl, &msg)
    if err != nil {
	fmt.Println(err)
        return err
    }

    return nil
}
