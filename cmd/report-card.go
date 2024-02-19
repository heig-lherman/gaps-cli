package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"lutonite.dev/gaps-cli/gaps"
	"lutonite.dev/gaps-cli/parser"
	"lutonite.dev/gaps-cli/util"
	"os"
	"strconv"
)

type ReportCardCmdOpts struct {
	format string
}

var (
	reportCardOpts = &ReportCardCmdOpts{}
	reportCardCmd  = &cobra.Command{
		Use:   "report-card",
		Short: "Allows to consult your report card",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := buildTokenClientConfiguration()

			action := gaps.NewReportCardAction(cfg)
			reports, err := action.FetchReportCard()
			util.CheckErr(err)

			if len(reports) == 0 {
				log.Error("No reports found for the given parameters")
				return nil
			}

			if reportCardOpts.format == "json" {
				return json.NewEncoder(os.Stdout).Encode(reports)
			}

			reportCardOpts.PrintReportCardTable(reports)
			return nil
		},
	}
)

func init() {
	reportCardCmd.Flags().StringVarP(&reportCardOpts.format, "format", "o", "table", "Output format (table, json)")

	rootCmd.AddCommand(reportCardCmd)
}

func (g *ReportCardCmdOpts) PrintReportCardTable(moduleReports []*parser.ModuleReport) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.Style().Options.SeparateRows = true
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 3, AutoMerge: true},
		{Number: 4, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 5, Align: text.AlignCenter, AlignHeader: text.AlignCenter, AlignFooter: text.AlignRight},
		{Number: 6, Align: text.AlignCenter, AlignHeader: text.AlignCenter, AlignFooter: text.AlignCenter},
	})

	t.AppendHeader(table.Row{"Module", "Credits", "Class", "Category", "Weight", "Grade"})

	for _, module := range moduleReports {
		moduleDesc := fmt.Sprintf("%s (%s)", module.Name, module.Identifier)
		if module.Year > 0 {
			moduleDesc += fmt.Sprintf(" - %d-%d", module.Year, module.Year+1)
		}

		for _, group := range module.Classes {
			groupDesc := fmt.Sprintf("%s (%s)", group.Name, group.Identifier)
			for _, grade := range group.Grades {
				t.AppendRow(table.Row{
					moduleDesc,
					module.Credits,
					groupDesc,
					grade.Name,
					fmt.Sprintf("%d%%", grade.Weight),
					grade.Grade,
				})
			}

			if group.Mean != "" {
				t.AppendRow(table.Row{
					moduleDesc,
					module.Credits,
					"",
					"",
					"",
					fmt.Sprintf("%s (W: %d)", group.Mean, group.Weight),
				}, table.RowConfig{AutoMerge: true})
			}
		}

		situation := text.Colors{text.FgGreen}.Sprint(module.Situation)
		t.AppendRow(table.Row{
			moduleDesc,
			module.Credits,
			situation,
			situation,
			situation,
			text.Colors{text.Bold, text.FgBlue}.Sprint(module.GlobalGrade),
		}, table.RowConfig{AutoMerge: true})

		t.AppendSeparator()
	}

	t.AppendFooter(table.Row{
		"",
		"",
		"",
		"",
		"WEIGHTED GPA",
		fmt.Sprintf("%.2f", computeGpa(moduleReports)),
	}, table.RowConfig{AutoMerge: true})

	t.Render()
}

func computeGpa(grades []*parser.ModuleReport) float64 {
	var totalCredits uint
	var totalPoints float64

	for _, module := range grades {
		if module.Situation != "Réussite" {
			continue
		}

		totalCredits += module.Credits
		gradeNumeric, _ := strconv.ParseFloat(module.GlobalGrade, 64)
		totalPoints += float64(module.Credits) * gradeNumeric
	}

	return totalPoints / float64(totalCredits)
}
