package main

import "time"

type chatMessageContent struct {
	Command string `json:"command"`
	From    struct {
		Mode string `json:"mode"`
		Nick string `json:"nick"`
	} `json:"from"`
	Highlight bool          `json:"highlight"`
	ID        int           `json:"id"`
	Params    []string      `json:"params"`
	Previews  []interface{} `json:"previews"`
	Self      bool          `json:"self"`
	Text      string        `json:"text"`
	Time      time.Time     `json:"time"`
	Type      string        `json:"type"`
	Users     []string      `json:"users"`
}

type chatMessage struct {
	Chan   int                `json:"chan"`
	Msg    chatMessageContent `json:"msg"`
	Unread int                `json:"unread"`
}

type channel struct {
	Name          string               `json:"name"`
	Type          string               `json:"type"`
	ID            int                  `json:"id"`
	Messages      []chatMessageContent `json:"messages"`
	TotalMessages int                  `json:"totalMessages"`
	Key           string               `json:"key"`
	Topic         string               `json:"topic"`
	State         int                  `json:"state"`
	FirstUnread   int                  `json:"firstUnread"`
	Unread        int                  `json:"unread"`
	Highlight     int                  `json:"highlight"`
	Users         []string             `json:"users"`
}

type network struct {
	UUID          string    `json:"uuid"`
	Name          string    `json:"name"`
	Nick          string    `json:"nick"`
	Channels      []channel `json:"channels"`
	ServerOptions struct {
		CHANTYPES []string `json:"CHANTYPES"`
		PREFIX    []string `json:"PREFIX"`
		NETWORK   string   `json:"NETWORK"`
	} `json:"serverOptions"`
	Status struct {
		Connected bool `json:"connected"`
		Secure    bool `json:"secure"`
	} `json:"status"`
}

type initMessage struct {
	Active   int       `json:"active"`
	Networks []network `json:"networks"`
	Token    string    `json:"token"`
}

func (i initMessage) NetworkByNameOrUUID(id string) *network {
	for _, n := range i.Networks {
		if n.Name == id || n.UUID == id {
			return &n
		}
	}

	return nil
}
