package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"lutonite.dev/gaps-cli/gaps"
	"lutonite.dev/gaps-cli/parser"
)

type AbsencesPeriod string

const (
	ALL        AbsencesPeriod = "all"
	ETE        AbsencesPeriod = "ete"
	SEMESTRE_1 AbsencesPeriod = "1"
	SEMESTRE_2 AbsencesPeriod = "2"
)

type AbsencesCmdOpts struct {
	format   string
	year     uint
	semester AbsencesPeriod
	minRate  uint
}

var (
	absencesOpts = &AbsencesCmdOpts{}
	absencesCmd  = &cobra.Command{
		Use:   "absences",
		Short: "Allows to consult your absences",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch absencesOpts.semester {
			case ALL, ETE, SEMESTRE_1, SEMESTRE_2: // valid
			default:
				return fmt.Errorf("invalid semester: %s. Must be one of: all, ete, 1, 2", absencesOpts.semester)
			}

			cfg := buildTokenClientConfiguration()
			absencesAction := gaps.NewAbsencesAction(cfg, absencesOpts.year)

			absences, err := absencesAction.FetchAbsences()
			if err != nil {
				return fmt.Errorf("couldn't fetch absences: %w", err)
			}

			if absencesOpts.format == "json" {
				return json.NewEncoder(os.Stdout).Encode(absences)
			}

			printAbsences(absences)
			return nil
		},
	}
)

func init() {
	absencesCmd.Flags().StringVarP(&absencesOpts.format, "format", "o", "table", "Output format (table, json)")
	absencesCmd.Flags().UintVarP(&absencesOpts.year, "year", "y", currentAcademicYear(),
		"Academic year (year at the start of the academic year, e.g. 2020 for 2020-2021 academic year)")
	absencesCmd.Flags().StringVarP((*string)(&absencesOpts.semester), "semester", "s", string(ALL),
		"Period to calculate absences for (all, ete, 1, 2)")
	absencesCmd.Flags().UintVarP(&absencesOpts.minRate, "rate", "r", 0, "Minimum rate to display")
	rootCmd.AddCommand(absencesCmd)
}

func printAbsences(absences *parser.AbsenceReport) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.Style().Options.SeparateRows = true
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignHeader: text.AlignCenter},
		{Number: 2, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 3, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 4, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
	})
	t.AppendHeader(table.Row{"Course", "Total", "Taux relatif", "Taux absolu"})

	for _, a := range absences.Courses {
		totalAbsence := a.Total - a.Justified
		selected, relativePresence, absolutePresence := calculateAbsences(&a)

		getColoredRate := func(rate float64) string {
			rateStr := fmt.Sprintf("%.2f%%", rate)
			switch {
			case rate >= 15:
				return text.Colors{text.FgRed, text.Bold}.Sprint(rateStr)
			case rate >= 8:
				return text.Colors{text.FgYellow}.Sprint(rateStr)
			default:
				return text.Colors{text.FgGreen}.Sprint(rateStr)
			}
		}

		if selected && absolutePresence >= float64(absencesOpts.minRate) {
			t.AppendRow(table.Row{
				a.Name,
				totalAbsence,
				getColoredRate(relativePresence),
				getColoredRate(absolutePresence),
			})
		}
	}
	t.Render()
}

func calculateAbsences(a *parser.CourseAbsence) (bool, float64, float64) {
	selected := false
	var relativePresence float64
	var absolutePresence float64
	switch absencesOpts.semester {
	case ETE:
		if selected = (a.Periods.Ete-a.Justified > 0); !selected {
			return selected, 0.0, 0.0
		}
		relativePresence = float64(a.Periods.Ete) / float64(a.RelativePeriods)
		absolutePresence = float64(a.Periods.Ete) / float64(a.AbsolutePeriods)

	case SEMESTRE_1:
		if selected = (a.Periods.Term1-a.Justified > 0) || (a.Periods.Term2-a.Justified > 0); !selected {
			return selected, 0.0, 0.0
		}
		relativePresence = float64(a.Periods.Term1+a.Periods.Term2-a.Justified) / float64(a.RelativePeriods)
		absolutePresence = float64(a.Periods.Term1+a.Periods.Term2-a.Justified) / float64(a.AbsolutePeriods)

	case SEMESTRE_2:
		if selected = (a.Periods.Term3-a.Justified > 0) || (a.Periods.Term4-a.Justified > 0); !selected {
			return selected, 0.0, 0.0
		}
		relativePresence = float64(a.Periods.Term3+a.Periods.Term4-a.Justified) / float64(a.RelativePeriods)
		absolutePresence = float64(a.Periods.Term3+a.Periods.Term4-a.Justified) / float64(a.AbsolutePeriods)

	default:
		selected = true
		relativePresence = float64(a.Total-a.Justified) / float64(a.RelativePeriods)
		absolutePresence = float64(a.Total-a.Justified) / float64(a.AbsolutePeriods)
	}
	return selected, relativePresence * 100.0, absolutePresence * 100.0
}
