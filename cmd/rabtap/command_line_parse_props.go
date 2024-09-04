// command line message properties parsing
// Copyright (C) 2024 Jan Delgado

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseMessageProperties fills a PropertiesOverride with the contents of
// the --props K=V map
func parseMessageProperties(args map[string]string) (PropertiesOverride, error) {
	props := PropertiesOverride{}
	clone := func(s string) *string { return &s }
	for k, v := range args {
		switch strings.ToLower(k) {
		case "contenttype":
			props.ContentType = clone(v)
		case "contentencoding":
			props.ContentEncoding = clone(v)
		case "deliverymode":
			var mode uint8
			switch strings.ToLower(v) {
			case "persistent":
				mode = 2
			case "transient":
				mode = 1
			default:
				return props, fmt.Errorf(`delivery mode must be "persistent" or "transient"`)
			}
			props.DeliveryMode = &mode
		case "priority":
			mode, err := strconv.Atoi(v)
			if err != nil {
				return props, fmt.Errorf("priority: %w", err)

			}
			if mode < 0 || mode > 255 {
				return props, fmt.Errorf("priority must be 0..255")
			}
			mode8 := uint8(mode)
			props.Priority = &mode8
		case "correlationid":
			props.CorrelationID = clone(v)
		case "replyto":
			props.ReplyTo = clone(v)
		case "expiration":
			// although te expiration is in ms, the attribute is a string
			props.Expiration = clone(v)
		case "messageid":
			props.MessageID = clone(v)
		case "timestamp":
			ts, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return props, fmt.Errorf("timestamp: %w", err)
			}
			props.Timestamp = &ts
		case "type":
			props.Type = clone(v)
		case "userid":
			props.UserID = clone(v)
		case "appid":
			props.AppID = clone(v)
		default:
			return props, fmt.Errorf("unknown property: %s", k)
		}
	}
	return props, nil
}
