package provider

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
)

// StartApplicationSessionResponse StartApplicationSessionResponse
type StartApplicationSessionResponse struct {
	XMLName               xml.Name `xml:"StartApplicationSessionPosResponse"`
	Xmlns                 string   `xml:"xmlns,attr"`
	SessionID             string   `xml:"sessionID"`
	ActualProtocolVersion string   `xml:"actualProtocolVersion"`
	ActualSessionDuration int      `xml:"actualSessionDuration"`
}

// StopApplicationSessionResponse StopApplicationSessionResponse
type StopApplicationSessionResponse struct {
	XMLName xml.Name `xml:"<StopApplicationSessionPosResponse"`
	Xmlns   string   `xml:"xmlns,attr"`
}

// Device Device
type Device struct {
	XMLName      xml.Name `xml:"device"`
	TypeOfNumber string   `xml:"typeOfNumber,attr"`
	MediaClass   string   `xml:"mediaClass,attr"`
	BitRate      string   `xml:"bitRate,attr"`
	ID           string   `xml:",chardata"`
}

// GetDeviceIDResponse GetDeviceIDResponse
type GetDeviceIDResponse struct {
	XMLName xml.Name `xml:"GetDeviceIdResponse"`
	Xmlns   string   `xml:"xmlns,attr"`
	Device  Device   `xml:"device"`
}

// MonitorStartResponse MonitorStartResponse
type MonitorStartResponse struct {
	XMLName           xml.Name `xml:"MonitorStartResponse"`
	MonitorCrossRefID string   `xml:"monitorCrossRefID"`
}

// MonitorStopResponse MonitorStopResponse
type MonitorStopResponse struct {
	XMLName xml.Name `xml:"MonitorStopResponse"`
	Xmlns   string   `xml:"xmlns,attr"`
}

// ResetApplicationSessionTimerResponse ResetApplicationSessionTimerResponse
type ResetApplicationSessionTimerResponse struct {
	XMLName               xml.Name `xml:"ResetApplicationSessionTimerPosResponse"`
	ActualSessionDuration int      `xml:"actualSessionDuration"`
}

// CSTAErrorCodeResponse CSTAErrorCodeResponse
type CSTAErrorCodeResponse struct {
	XMLName                        xml.Name `xml:"CSTAErrorCode"`
	Xmlns                          string   `xml:"xmlns,attr"`
	Operation                      string   `xml:"operation"`
	Security                       string   `xml:"security"`
	StateIncompatibility           string   `xml:"stateIncompatibility"`
	SystemResourceAvailibility     string   `xml:"systemResourceAvailibility"`
	SubscribedResourceAvailability string   `xml:"subscribedResourceAvailability"`
	PerformanceManagement          string   `xml:"performanceManagement"`
	PrivateData                    string   `xml:"privateData"`
	Unspecified                    string   `xml:"unspecified"`
}

//StartApplicationSessionMessage StartApplicationSessionMessage
func StartApplicationSessionMessage(appName string, user string, password string, sessionCleanupDelay string, requestedSessionDuration string) string {
	var message bytes.Buffer
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.WriteString("<StartApplicationSession xmlns=\"http://www.ecma-international.org/standards/ecma-354/appl_session\">")
	message.WriteString("<applicationInfo>")
	message.WriteString("<applicationID>" + appName + "</applicationID>")
	message.WriteString("<applicationSpecificInfo>")
	message.WriteString("<ns1:SessionLoginInfo ")
	message.WriteString("xmlns:ns1=\"http://www.pbxnsip.com/schemas/csta\" ")
	message.WriteString("xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" ")
	message.WriteString("xsi:type=\"ns1:SessionLoginInfo\">")
	message.WriteString("<ns1:userName>" + user + "</ns1:userName>")
	message.WriteString("<ns1:password>" + password + "</ns1:password>")
	message.WriteString("<ns1:sessionCleanupDelay>" + sessionCleanupDelay + "</ns1:sessionCleanupDelay>")
	message.WriteString("</ns1:SessionLoginInfo>")
	message.WriteString("</applicationSpecificInfo>")
	message.WriteString("</applicationInfo>")
	message.WriteString("<requestedProtocolVersions>")
	message.WriteString("<protocolVersion>http://www.ecma-international.org/standards/ecma-323/csta/ed3/priv5</protocolVersion>")
	message.WriteString("</requestedProtocolVersions>")
	message.WriteString("<requestedSessionDuration>" + requestedSessionDuration + "</requestedSessionDuration>")
	message.WriteString("</StartApplicationSession>")
	return message.String()
}

// ResetApplicationSessionTimerMessage ResetApplicationSessionTimerMessage
func ResetApplicationSessionTimerMessage(sessionID string) string {
	var message bytes.Buffer
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.WriteString("<ResetApplicationSessionTimer xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" xmlns=\"http://www.ecma-international.org/standards/ecma- 354/appl_session\">")
	message.WriteString("<sessionID>")
	message.WriteString(sessionID)
	message.WriteString("</sessionID>")
	message.WriteString("<requestedSessionDuration>180</requestedSessionDuration>")
	message.WriteString("</ResetApplicationSessionTimer>")
	return message.String()
}

// GetDeviceIDMessage GetDeviceIDMessage
func GetDeviceIDMessage(callServerIP string, extension string) string {
	var message bytes.Buffer
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.WriteString("<GetDeviceId xmlns=\"http://www.pbxnsip.com/schemas/csta\">")
	message.WriteString("<switchName>" + callServerIP + "</switchName>")
	message.WriteString("<extension>" + extension + "</extension>")
	message.WriteString("</GetDeviceId>")
	return message.String()
}

// MonitorVDNStartMessage MonitorVDNStartMessage
func MonitorVDNStartMessage(deviceID string) string {
	var message bytes.Buffer
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.WriteString("<MonitorStart xmlns=\"http://www.ecma-international.org/standards/ecma-323/csta/ed3\">")
	message.WriteString("<monitorObject>")
	message.WriteString("<deviceObject typeOfNumber=\"other\" mediaClass=\"notKnown\" bitRate=\"constant\">")
	message.WriteString(deviceID)
	message.WriteString("</deviceObject>")
	message.WriteString("</monitorObject>")
	message.WriteString("<requestedMonitorFilter>")
	message.WriteString("<callcontrol>")
	message.WriteString("<callCleared>true</callCleared>")
	message.WriteString("<conferenced>true</conferenced>")
	message.WriteString("<connectionCleared>true</connectionCleared>")
	message.WriteString("<delivered>true</delivered>")
	message.WriteString("<diverted>true</diverted>")
	message.WriteString("<established>true</established>")
	message.WriteString("<failed>true</failed>")
	message.WriteString("<held>true</held>")
	message.WriteString("<networkReached>true</networkReached>")
	message.WriteString("<originated>true</originated>")
	message.WriteString("<queued>true</queued>")
	message.WriteString("<retrieved>true</retrieved>")
	message.WriteString("<serviceInitiated>true</serviceInitiated>")
	message.WriteString("<transferred>true</transferred>")
	message.WriteString("</callcontrol>")
	message.WriteString("<callAssociated>")
	message.WriteString("<callInformation>true</callInformation>")
	message.WriteString("<charging>true</charging>")
	message.WriteString("<digitsGenerated>true</digitsGenerated>")
	message.WriteString("<telephonyTonesGenerated>true</telephonyTonesGenerated>")
	message.WriteString("<serviceCompletionFailure>true</serviceCompletionFailure>")
	message.WriteString("</callAssociated>")
	message.WriteString("<logicalDeviceFeature />")
	message.WriteString("</requestedMonitorFilter>")
	message.WriteString("<monitorType>call</monitorType>")
	message.WriteString("<extensions>")
	message.WriteString("<privateData>")
	message.WriteString("<private>")
	message.WriteString("<Events xmlnssmiliesi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlnssmiliesd=\"http://www.w3.org/2001/XMLSchema\" xmlns=\"\"> ")
	message.WriteString("<invertFilter xmlns=\"hhttp://www.pbxnsip.com/schemas/csta\">true</invertFilter>")
	message.WriteString("<callControlPrivate xmlns=\"hhttp://www.pbxnsip.com/schemas/csta\"> ")
	message.WriteString("<enteredDigits>true</enteredDigits>")
	message.WriteString("</callControlPrivate>")
	message.WriteString("</Events>")
	message.WriteString("</private>")
	message.WriteString("</privateData>")
	message.WriteString("</extensions>")
	message.WriteString("</MonitorStart>")
	return message.String()
}

// MonitorSkillStartMessage MonitorSkillStartMessage
func MonitorSkillStartMessage(deviceID string) string {
	var message bytes.Buffer
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.WriteString("<MonitorStart xmlns=\"http://www.ecma-international.org/standards/ecma-323/csta/ed3\">")
	message.WriteString("<monitorObject>")
	message.WriteString("<deviceObject typeOfNumber=\"other\" mediaClass=\"notKnown\" bitRate=\"constant\">")
	message.WriteString(deviceID)
	message.WriteString("</deviceObject>")
	message.WriteString("</monitorObject>")
	message.WriteString("<requestedMonitorFilter>")
	message.WriteString("<logicalDeviceFeature />")
	message.WriteString("</requestedMonitorFilter>")
	message.WriteString("</MonitorStart>")
	return message.String()
}

// MonitorStopMessage MonitorStopMessage
func MonitorStopMessage(monitorID string) string {
	var message bytes.Buffer
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.WriteString("<MonitorStop xmlns=\"hhttp://www.pbxnsip.com/schemas/csta\">")
	message.WriteString("<monitorCrossRefID>")
	message.WriteString(monitorID)
	message.WriteString("</monitorCrossRefID>")
	message.WriteString("</MonitorStop>")
	return message.String()
}

// StopAppSessionMessage StopAppSessionMessage
func StopAppSessionMessage(sessionID string) string {
	var message bytes.Buffer
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.WriteString("<StopApplicationSession xmlns=\"http://www.ecma-international.org/standards/ecma-354/appl_session\">")
	message.WriteString("<sessionID>")
	message.WriteString(sessionID)
	message.WriteString("</sessionID>")
	message.WriteString("<sessionEndReason>")
	message.WriteString("<definedEndReason>normal</definedEndReason>")
	message.WriteString("</sessionEndReason>")
	message.WriteString("</StopApplicationSession>")
	return message.String()
}

// ParseMessageResponse ParseMessageResponse
func ParseMessageResponse(data string, response interface{}) error {
	if err := xml.Unmarshal([]byte(data), &response); err != nil {
		fmt.Printf("error: %v\n", err)
		var failure CSTAErrorCodeResponse
		if err := xml.Unmarshal([]byte(data), &failure); err != nil {
			fmt.Printf("error: %v\n", err)
			return err
		}
		fmt.Printf("CSTAErrorCodeResponse: %v\n", failure)
		return errors.New(failure.Operation)
	}
	return nil
}
