package cmd

import (
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"lutonite.dev/gaps-cli/gaps"
)

type Location uint8

const (
	Cheseaux Location = iota
	StRoch
)

func (l Location) String() string {
	switch l {
	case Cheseaux:
		return "Cheseaux"
	case StRoch:
		return "St-Roch"
	default:
		return "Unknown"
	}
}

// Semesters are trimesters on GAPS. So for semester 2, it is trimester 3
func correctSemester(semester uint) uint {
	if semester == 2 {
		return 3
	}
	return semester
}

const timeLayout = "15:04"

var (
	weekdayFlag   string
	startTimeFlag string
	endTimeFlag   string
	siteFlag      string
	semesterFlag  uint

	findFreeRoomsCmd = &cobra.Command{
		Use:   "find-free-rooms",
		Short: "Finds free rooms based on specified criteria",
		Run: func(cmd *cobra.Command, args []string) {
			freeStart, err := time.Parse(timeLayout, startTimeFlag)
			if err != nil {
				log.WithError(err).Fatalf("Failed to parse start time: %s", startTimeFlag)
			}

			freeEnd, err := time.Parse(timeLayout, endTimeFlag)
			if err != nil {
				log.WithError(err).Errorf("Failed to parse end time: %s", endTimeFlag)
			}

			var targetSite Location
			switch siteFlag {
			case "Cheseaux":
				targetSite = Cheseaux
			case "St-Roch":
				targetSite = StRoch
			default:
				log.Fatalf("Invalid site specified: %s. Must be 'Cheseaux' or 'St-Roch'", siteFlag)
				return
			}

			validWeekdays := map[string]bool{
				"Monday": true, "Tuesday": true, "Wednesday": true, "Thursday": true, "Friday": true,
			}
			if !validWeekdays[weekdayFlag] {
				log.Errorf("Invalid weekday specified: %s", weekdayFlag)
				return
			}

			if semesterFlag < 0 || semesterFlag > 2 {
				log.Errorf("Invalid semester specified: %d. Must be between 1 and 3", semesterFlag)
				return
			}

			cfg := buildTokenClientConfiguration()
			registryAction := gaps.NewRegistryAction(cfg, currentAcademicYear())

			registry, err := registryAction.FetchRegistry()
			if err != nil {
				log.Fatal(err)
			}

			var freeRooms []string

			weekday := weekdayFlag
			semester := semesterFlag

			fmt.Printf("Searching for free rooms at %s on %s between %s and %s for semester %d...",
				targetSite, weekday, freeStart.Format(timeLayout), freeEnd.Format(timeLayout), semester)

			freeStartMinutes := freeStart.Hour()*60 + freeStart.Minute()
			freeEndMinutes := freeEnd.Hour()*60 + freeEnd.Minute()
			for _, regRoom := range registry.Rooms {
				var currentRoomSite Location
				switch regRoom.Name[0] {
				case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K':
					currentRoomSite = Cheseaux
				default:
					currentRoomSite = StRoch
				}

				if currentRoomSite != targetSite {
					continue
				}

				roomScheduleAction := gaps.NewRoomScheduleAction(cfg, currentAcademicYear(), correctSemester(semester), regRoom.ID)
				calendar, err := roomScheduleAction.FetchSchedule()
				if err != nil {
					log.WithError(err).Warnf("Failed to fetch schedule for room %s (ID: %d)", regRoom.Name, regRoom.ID)
					continue
				}

				isAvailable := true
				for _, event := range calendar.Events() {
					start, err := parseDateTime(event.GetProperty("DTSTART").Value)
					if err != nil {
						log.WithError(err).Warnf("Failed to parse event start time for room %s", regRoom.Name)
						continue
					}

					end, err := parseDateTime(event.GetProperty("DTEND").Value)
					if err != nil {
						log.WithError(err).Warnf("Failed to parse event end time for room %s", regRoom.Name)
						continue
					}

					if weekday != start.Weekday().String() {
						continue
					}

					eventStartMinutes := start.Hour()*60 + start.Minute()
					eventEndMinutes := end.Hour()*60 + end.Minute()

					if eventStartMinutes < freeEndMinutes && eventEndMinutes > freeStartMinutes {
						isAvailable = false
						break
					}
				}

				if isAvailable {
					freeRooms = append(freeRooms, regRoom.Name)
				}
			}

			sort.Strings(freeRooms)
			if len(freeRooms) > 0 {
				fmt.Printf("\rFound %d free rooms at %s on %s between %s and %s for semester %d:       \n",
					len(freeRooms), targetSite, weekday, freeStart.Format(timeLayout), freeEnd.Format(timeLayout), semester)
				for _, roomName := range freeRooms {
					fmt.Printf("- %s\n", roomName)
				}
			} else {
				fmt.Printf("\nNo free rooms found at %s on %s between %s and %s for semester %d.\n",
					targetSite, weekday, freeStart.Format(timeLayout), freeEnd.Format(timeLayout), semester)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(findFreeRoomsCmd)
	findFreeRoomsCmd.Flags().StringVarP(&weekdayFlag, "weekday", "w", "Monday", "Day of the week (e.g., Monday)")
	findFreeRoomsCmd.Flags().StringVarP(&startTimeFlag, "start-time", "s", "08:00", "Start time in HH:MM format")
	findFreeRoomsCmd.Flags().StringVarP(&endTimeFlag, "end-time", "e", "10:00", "End time in HH:MM format")
	findFreeRoomsCmd.Flags().StringVarP(&siteFlag, "site", "l", "Cheseaux", "Site location ('Cheseaux' or 'St-Roch')")
	findFreeRoomsCmd.Flags().UintVarP(&semesterFlag, "semester", "m", 1, "Semester number (0, 1, 2), 0 is summer")

	findFreeRoomsCmd.MarkFlagRequired("weekday")
	findFreeRoomsCmd.MarkFlagRequired("start-time")
	findFreeRoomsCmd.MarkFlagRequired("end-time")
	findFreeRoomsCmd.MarkFlagRequired("semester")
}
