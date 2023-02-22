package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ch "lutonite.dev/gaps-cli/cal"
	"lutonite.dev/gaps-cli/gaps"
	"lutonite.dev/gaps-cli/parser"
	"lutonite.dev/gaps-cli/util"
	"os"
	"strconv"
	"strings"
	"time"
)

type GradesCmdOpts struct {
	format string
	year   string
	class  string
}

var (
	gradesOpts = &GradesCmdOpts{}
	gradesCmd  = &cobra.Command{
		Use:   "grades",
		Short: "Allows to consult your grades",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := buildTokenClientConfiguration()

			var classGrades []*parser.ClassGrades
			for _, sYear := range strings.Split(gradesOpts.year, ",") {
				year, err := strconv.ParseUint(sYear, 10, 32)
				util.CheckErr(err)
				grades := gaps.NewGradesAction(cfg, uint(year))
				grades.ClassFilter = gradesOpts.class
				res, err := grades.FetchGrades()
				util.CheckErr(err)
				classGrades = append(classGrades, res...)
			}

			if len(classGrades) == 0 {
				log.Error("No grades found for the given parameters")
				return nil
			}

			if gradesOpts.format == "json" {
				return json.NewEncoder(os.Stdout).Encode(classGrades)
			}

			gradesOpts.PrintGradesTable(classGrades)
			return nil
		},
	}
)

func init() {
	gradesCmd.Flags().StringVarP(&gradesOpts.format, "format", "o", "table", "Output format (table, json)")
	gradesCmd.Flags().StringVar(&gradesOpts.class, "class", "", "Get grades for specific class")
	gradesCmd.Flags().StringVarP(
		&gradesOpts.year, "year", "y", gradesOpts.defaultYear(),
		"Academic year (year at the start of the academic year, e.g. 2020 for 2020-2021 academic year)",
	)

	rootCmd.AddCommand(gradesCmd)
}

func (*GradesCmdOpts) defaultYear() string {
	actual, _ := ch.BettagMontag.Calc(time.Now().Year())
	if time.Now().Before(actual) {
		return fmt.Sprintf("%d", time.Now().Year()-1)
	}

	return fmt.Sprintf("%d", time.Now().Year())
}

func (g *GradesCmdOpts) PrintGradesTable(classGrades []*parser.ClassGrades) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.Style().Options.SeparateRows = true
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true},
		{Number: 3, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 4, AlignHeader: text.AlignCenter, WidthMin: 30},
		{Number: 5, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 6, Align: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 7, Align: text.AlignCenter, AlignHeader: text.AlignCenter, AlignFooter: text.AlignCenter},
	})

	t.AppendHeader(table.Row{"Class", "Group", "Date", "Description", "Class Mean", "Weight", "Grade"})

	for _, classGrade := range classGrades {
		classDesc := classGrade.Name
		if classGrade.HasExam {
			classDesc += " (E)"
		}
		classDesc = fmt.Sprintf("%s - mean %s", classDesc, classGrade.GlobalMean)

		for _, group := range classGrade.GradeGroups {
			for _, grade := range group.Grades {
				t.AppendRow(table.Row{
					classDesc,
					group.Name,
					grade.Date.Format("02.01.2006"),
					grade.Description,
					grade.ClassMean,
					fmt.Sprintf("%.1f%%", grade.Weight),
					grade.Grade,
				})
			}
			t.AppendRow(table.Row{
				classDesc,
				"",
				"",
				"",
				"",
				"",
				fmt.Sprintf("%s (W: %d%%)", group.Mean, group.Weight),
			}, table.RowConfig{AutoMerge: true})
		}

		t.AppendSeparator()
	}

	t.Render()
}
