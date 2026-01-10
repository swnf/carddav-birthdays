package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/SimonSchneider/goslu/date"
)

type Birthday struct {
	UID      string
	Date     date.Date
	FullName string
}

func parseBirthdayVCard(vcard string) *Birthday {
	lines := strings.Split(vcard, "\n")
	var name, fullName, uid string
	var birthdayDate *time.Time

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse N (Name components)
		if strings.HasPrefix(line, "N:") {
			name = strings.TrimPrefix(line, "N:")
		}

		// Parse FN (Full Name)
		if strings.HasPrefix(line, "FN:") {
			fullName = strings.TrimPrefix(line, "FN:")
		}
		if strings.HasPrefix(line, "UID:") {
			uid = strings.TrimPrefix(line, "UID:")
		}

		// Parse BDAY (Birthday) - handle BDAY:, BDAY;VALUE=date:, and BDAY;X-APPLE-OMIT-YEAR=1604: formats
		if strings.HasPrefix(line, "BDAY") {
			var bdayStr string

			// Handle BDAY;VALUE=date: format
			if strings.Contains(line, ";VALUE=date:") {
				parts := strings.Split(line, ";VALUE=date:")
				if len(parts) == 2 {
					bdayStr = parts[1]
				}
			} else if strings.Contains(line, ";X-APPLE-OMIT-YEAR=1604:") {
				parts := strings.Split(line, ";X-APPLE-OMIT-YEAR=1604:")
				if len(parts) == 2 {
					bdayStr = strings.Replace(parts[1], "1604", "0000", 1)
				}
			} else if strings.HasPrefix(line, "BDAY:") {
				// Handle simple BDAY: format
				bdayStr = strings.TrimPrefix(line, "BDAY:")
			}

			if bdayStr != "" {
				// Remove timezone info if present
				if idx := strings.Index(bdayStr, "T"); idx != -1 {
					bdayStr = bdayStr[:idx]
				}

				// Try different date formats
				// --0415 is valid if there is no year
				formats := []string{"20060102", "2006-01-02", "2006/01/02", "060102", "--0102"}
				for _, format := range formats {
					if date, err := time.Parse(format, bdayStr); err == nil {
						birthdayDate = &date
						break
					}
				}
				if birthdayDate == nil {
					fmt.Printf("failed to parse birthday date: %s\n", bdayStr)
				}
			}
		}
	}

	// Only return birthday if we have both name and date
	if birthdayDate != nil && (name != "" || fullName != "") {
		displayName := fullName
		if displayName == "" {
			displayName = name
		}

		return &Birthday{
			UID:      uid,
			FullName: displayName,
			Date:     date.FromTime(*birthdayDate),
		}
	}

	return nil
}

func generateBirthdayIcs(birthdays []Birthday, today date.Date) string {
	var sb strings.Builder

	// iCalendar header
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//CardDAV Birthdays//EN\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")

	// Add each birthday as a recurring event
	for _, birthday := range birthdays {
		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString("UID:" + fmt.Sprintf("%s-birthday-%d", strings.ReplaceAll(birthday.FullName, " ", ""), birthday.Date.ToStdTime().Year()) + "\r\n")
		if birthday.Date.ToStdTime().Year() == 0 {
			sb.WriteString("DTSTART;VALUE=DATE:1900" + birthday.Date.ToStdTime().Format("0102") + "\r\n")
		} else {
			sb.WriteString("DTSTART;VALUE=DATE:" + birthday.Date.ToStdTime().Format("20060102") + "\r\n")
		}
		sb.WriteString("RRULE:FREQ=YEARLY\r\n")
		sb.WriteString("SUMMARY:" + birthday.FullName + "'s Birthday\r\n")
		sb.WriteString("DESCRIPTION:" + birthday.FullName + " born on " + birthday.Date.String() + "\r\n")
		sb.WriteString("TRANSP:TRANSPARENT\r\n")
		sb.WriteString("END:VEVENT\r\n")
	}

	// iCalendar footer
	sb.WriteString("END:VCALENDAR\r\n")

	return sb.String()
}
