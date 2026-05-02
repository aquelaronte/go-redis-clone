package parser

import "strconv"

// readUntilNextCRLF reads a byte slice until it encounters the next CRLF sequence.
//
// It takes a byte slice called `received` as input, which contains the message.
// The function iterates over the `received` byte slice and checks if the current byte and the next byte are equal to '\r' and '\n' respectively.
// If a CRLF sequence is found, it returns a new byte slice containing the value until the CRLF sequence, and the remaining byte slice after the CRLF sequence.
// The CRLF sequence is skipped by excluding the '\r' and '\n' bytes from the returned byte slices.
//
// If there are no CRLF sequences found in the `received` byte slice, it means that the entire message is a remaining, so the function returns `nil` for the value and the entire `received` byte slice as the remaining.
//
// The function returns two byte slices: the value until the CRLF sequence, and the remaining byte slice after the CRLF sequence.
func readUntilNextCRLF(received []byte) ([]byte, []byte) {
	if len(received) == 0 {
		return nil, nil
	}

	for i := range received {
		// since we check received[i+1] in the next conditional,
		// then we must break in the last index
		if i == len(received)-1 {
			break
		}

		if received[i] == '\r' && received[i+1] == '\n' {
			value := received[:i]
			remaining := received[i+2:] // skip next \r\n bytes

			return value, remaining
		}
	}

	// If there are no CRLF, then the entire message is a remaining
	return nil, received
}

// retrieveLength retrieves the length of a message from a byte slice.
//
// It takes a byte slice called `received` as input, which contains the message.
// The function first calls the `readUntilNextCRLF` function to read the first part of the message until the next CRLF sequence is found.
// If the first part is `nil`, it means that there was no valid data found in the `received` byte slice, so the function returns `0`, `remaining`, and `nil` as the error.
//
// Otherwise, it converts the first part to a string and then uses `strconv.Atoi` to parse the string into an integer.
// The parsed length is returned along with `remaining` and any error that occurred during parsing.
func retrieveLength(received []byte) (int, []byte, error) {
	firstPart, remaining := readUntilNextCRLF(received)

	if len(firstPart) == 0 {
		return 0, remaining, nil
	}

	parsedLength, err := strconv.Atoi(string(firstPart[1:]))

	return parsedLength, remaining, err
}
