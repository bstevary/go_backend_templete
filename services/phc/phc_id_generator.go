package phc

import (
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

func GeneratePHCID() string {
	now := time.Now()

	// Generates a unique PHC ID using the current date, time, and a random component.
	// The format is as follows:
	// - The day and month are converted to a custom hexadecimal representation via customHex.
	// - The year is represented by its last two digits, zero-padded to ensure two characters.
	// - The hour is also converted to a custom hexadecimal representation.
	// - The minute and second are represented in uppercase hexadecimal, zero-padded.
	// - A random component is added at the end, represented as a zero-padded hexadecimal number in the range [0, 4095].
	// This results in a string that uniquely identifies a PHC (Primary Health Care) record with a timestamp and random component.
	return fmt.Sprintf("%s%s%02d%s%02X%02X%03X",
		customHex(now.Day()),        // Day in custom hex
		customHex(int(now.Month())), // Month in custom hex
		now.Year()%100,              // Last two digits of the year
		customHex(now.Hour()),       // Hour in custom hex
		now.Minute(),                // Minute in hex, zero-padded
		now.Second(),                // Second in hex, zero-padded
		rand.Intn(4095),             // Random component in hex, zero-padded
	)
}

func customHex(value int) string {
	// Guard clause for values outside the 0-35 range
	if value < 0 || value > 35 {
		return "0"
	}
	// Direct conversion for 0-9
	if value <= 9 {
		return string('0' + value)
	}
	// Direct conversion for 10-35 to A-Z
	return string('A' + (value - 10))
}

func GenerateHospitalID() string {
	return fmt.Sprintf("HS%02d%03X", time.Now().Year()%100, rand.Intn(4096))
}
