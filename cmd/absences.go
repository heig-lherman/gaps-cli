package cmd

import (
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
	format string
	year   string
	period AbsencesPeriod
}

var (
	absencesOpts = &AbsencesCmdOpts{}
	absencesCmd  = &cobra.Command{
		Use:   "absences",
		Short: "Allows to consult your absences",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch absencesOpts.period {
			case ALL, ETE, SEMESTRE_1, SEMESTRE_2:
				// valid
			default:
				return fmt.Errorf("invalid period: %s. Must be one of: all, ete, semestre1, semestre2", absencesOpts.period)
			}

			cfg := buildTokenClientConfiguration()
			absencesAction := gaps.NewAbsencesAction(cfg, currentAcademicYear())

			absences, err := absencesAction.FetchAbsences()
			if err != nil {
				return fmt.Errorf("couldn't fetch absences: %w", err)
			}
			printAbsences(absences)
			return nil
		},
	}
)

func init() {
	absencesCmd.Flags().StringVarP(&absencesOpts.format, "format", "o", "table", "Output format (table, json)")
	absencesCmd.Flags().StringVarP(
		&absencesOpts.year, "year", "y", absencesOpts.defaultYear(),
		"Academic year (year at the start of the academic year, e.g. 2020 for 2020-2021 academic year)",
	)
	absencesCmd.Flags().StringVarP(
		(*string)(&absencesOpts.period),
		"period",
		"p",
		string(ALL),
		"Period to calculate absences for (all, ete, semestre1, semestre2)",
	)
	rootCmd.AddCommand(absencesCmd)
}

func (*AbsencesCmdOpts) defaultYear() string {
	return fmt.Sprintf("%d", currentAcademicYear())
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
		relativePresence, absolutePresence := calculateAbsences(&a)

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

		selected := false
		switch absencesOpts.period {
		case ETE:
			selected = (a.Periods.Ete > 0)
		case SEMESTRE_1:
			selected = (a.Periods.Term1 > 0) || (a.Periods.Term2 > 0)
		case SEMESTRE_2:
			selected = (a.Periods.Term3 > 0) || (a.Periods.Term4 > 0)
		default:
			selected = true
		}

		if selected {
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

func calculateAbsences(a *parser.CourseAbsence) (float64, float64) {
	var relativePresence float64
	var absolutePresence float64
	switch absencesOpts.period {
	case ETE:
		if a.Periods.Ete != 0 {
			relativePresence = float64(a.Periods.Ete) / float64(a.RelativePeriods)
			absolutePresence = float64(a.Periods.Ete) / float64(a.AbsolutePeriods)
		} else {
			return 0, 0
		}
	case SEMESTRE_1:
		relativePresence = float64(a.Periods.Term1+a.Periods.Term2) / float64(a.RelativePeriods)
		absolutePresence = float64(a.Periods.Term1+a.Periods.Term2) / float64(a.AbsolutePeriods)
	case SEMESTRE_2:
		relativePresence = float64(a.Periods.Term3+a.Periods.Term4) / float64(a.RelativePeriods)
		absolutePresence = float64(a.Periods.Term3+a.Periods.Term4) / float64(a.AbsolutePeriods)
	default:
		relativePresence = float64(a.Total) / float64(a.RelativePeriods)
		absolutePresence = float64(a.Total) / float64(a.AbsolutePeriods)
	}
	return relativePresence * 100.0, absolutePresence * 100.0
}
