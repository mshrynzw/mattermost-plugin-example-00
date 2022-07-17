package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	if strings.Index(post.Message, "https://qiita.com/") == -1 {
		return post, ""
	}

	p.API.LogDebug("Qiita link is detected.")
	post.Message = fmt.Sprintf("%s #Qiita", post.Message)
	post.Hashtags = fmt.Sprintf("%s #Qiita", post.Hashtags)
	return post, ""
}
