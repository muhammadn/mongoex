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
    if slackWebhookUrl == "" {
            return nil
    }

    attachment := slack.Attachment{
	Fallback:      message,
	AuthorName:    "mongobot",
	AuthorIcon:    "https://avatars2.githubusercontent.com/u/652790",
	Text:          message,
	Footer:        "mongoex",
	FooterIcon:    "https://platform.slack-edge.com/img/default_application_icon.png",
	Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
    }
    msg := slack.WebhookMessage{
	Attachments: []slack.Attachment{attachment},
    }

    err := slack.PostWebhook(slackWebhookUrl, &msg)
    if err != nil {
	fmt.Println(err)
        return err
    }

    return nil
}
