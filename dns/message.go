package dns

import (
	"fmt"
)

// Message format:
// All communications inside of the domain protocol are carried in a single
// format called a message.  The top level format of message is divided
// into 5 sections (some of which are empty in certain cases) shown below:

//     +---------------------+
//     |        Header       |
//     +---------------------+
//     |       Question      | the question for the name server
//     +---------------------+
//     |        Answer       | RRs answering the question
//     +---------------------+
//     |      Authority      | RRs pointing toward an authority
//     +---------------------+
//     |      Additional     | RRs holding additional information
//     +---------------------+

type Message struct {
	Header      Header
	Questions   []Question
	Answers     []ResourceRecord
	NameServers []ResourceRecord
	Additionals []ResourceRecord
}

const MaxDNSMessageSizeOverUDP = 512
const MaxDNSMessageSize = 4096

// DecodeMessage parses DNS message data and returns a Message structure.
//
// Parameters:
//   - data: The DNS message in a byte slice.
//
// Returns:
//   - *Message: The decoded DNS message in a structure.
//   - error: If the message is invalid or decoding fails.
func DecodeMessage(data []byte) (Message, error) {
	reader := &dnsReader{data: data}

	header, err := reader.readHeader()
	if err != nil {
		return Message{}, invalidMessageError(err.Error())
	}
	if reader.offset != DNSHeaderLength {
		return Message{}, invalidMessageError("invalid offset after reading header")

	}

	questions, err := reader.readQuestions(header.QuestionCount)
	if err != nil {
		return Message{}, invalidMessageError(fmt.Sprintf("question section: %s", err.Error()))
	}

	answers, err := reader.readResourceRecords(header.AnswerRRCount)
	if err != nil {
		return Message{}, invalidMessageError(fmt.Sprintf("answer section: %s", err.Error()))
	}

	nameServers, err := reader.readResourceRecords(header.NameserverRRCount)
	if err != nil {
		return Message{}, invalidMessageError(fmt.Sprintf("authority section: %s", err.Error()))
	}

	additionals, err := reader.readResourceRecords(header.AdditionalRRCount)
	if err != nil {
		return Message{}, invalidMessageError(fmt.Sprintf("additional section: %s", err.Error()))
	}

	return Message{
		Header:      header,
		Questions:   questions,
		Answers:     answers,
		NameServers: nameServers,
		Additionals: additionals,
	}, nil
}

// EncodeMessage converts a Message structure into DNS message bytes.
//
// Parameters:
//   - msg: A pointer to a Message structure to encode.
//
// Returns:
//   - []byte: The encoded DNS message bytes.
//   - error: If encoding fails.
func EncodeMessage(message Message) ([]byte, error) {
	writer := &dnsWriter{
		data:   make([]byte, DNSHeaderLength),
		offset: 0,
	}

	writer.writeHeader(message)

	writer.writeQuestions(message.Questions)
	writer.writeResourceRecords(message.Answers)
	writer.writeResourceRecords(message.NameServers)
	writer.writeResourceRecords(message.Additionals)

	return writer.data, nil
}
