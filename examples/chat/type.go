package main

import "encoding/json"

type Sms struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Content string `json:"content"`
}

func parseSms(msgInterface interface{}) (*Sms, error) {
	sms := &Sms{}
	msgBytes, err := json.Marshal(msgInterface)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(msgBytes, sms)
	if err != nil {
		return nil, err
	}
	return sms, nil
}
