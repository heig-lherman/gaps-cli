package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/r3labs/diff/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"lutonite.dev/gaps-cli/gaps"
	"lutonite.dev/gaps-cli/notifier"
	"lutonite.dev/gaps-cli/parser"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ScraperCommand struct {
	apiUrl string
	apiKey string

	interval int

	historyFile string
}

type scraperResult map[string]map[string]*scraperGrade
type scraperGrade struct {
	Course      string    `json:"course" diff:"-"`
	Type        string    `json:"type" diff:"-"`
	Description string    `json:"description" diff:"Description, identifier"`
	Date        time.Time `json:"date" diff:"-"`
	Weight      float32   `json:"weight" diff:"-"`
	Grade       string    `json:"grade"`
	ClassMean   string    `json:"classMean"`
}

var (
	scraperOpts = &ScraperCommand{}
	scraperCmd  = &cobra.Command{
		Use:   "scraper",
		Short: "Runs a scraper for grades for the distributed Discord notifications API",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Refreshing token")
			refreshToken(
				defaultViper.GetString(UsernameViperKey.Key()),
				credentialsViper.GetString(PasswordViperKey.Key()),
			)

			log.Info("Starting scraper thread")

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			ticker := time.NewTicker(time.Duration(scraperOpts.interval) * time.Second)

			defer ticker.Stop()
			for {
				select {
				case <-c:
					log.Info("Received interrupt, exiting")
					return nil
				case <-ticker.C:
					if err := scraperOpts.runScraper(); err != nil {
						log.WithError(err).Error("Failed to run scraper")
					}
				}
			}
		},
	}
)

func init() {
	scraperCmd.Flags().StringP(UsernameViperKey.Flag(), "u", "", "einet aai username")
	defaultViper.BindPFlag(UsernameViperKey.Key(), scraperCmd.Flags().Lookup(UsernameViperKey.Flag()))

	scraperCmd.Flags().StringP(PasswordViperKey.Flag(), "p", "", "einet aai password")
	credentialsViper.BindPFlag(PasswordViperKey.Key(), scraperCmd.Flags().Lookup(PasswordViperKey.Flag()))

	scraperCmd.Flags().StringVar(&scraperOpts.historyFile, GradesHistoryFileViperKey.Flag(), "", "history file (default is $HOME/.config/gaps-cli/grades-history.json)")
	defaultViper.BindPFlag(GradesHistoryFileViperKey.Key(), scraperCmd.Flags().Lookup(GradesHistoryFileViperKey.Flag()))
	defaultViper.SetDefault(GradesHistoryFileViperKey.Key(), getConfigDirectory()+"/gaps-cli/grades-history.json")

	scraperCmd.Flags().StringVarP(&scraperOpts.apiUrl, ScraperApiUrlViperKey.Flag(), "U", "", "Notifier API URL")
	defaultViper.BindPFlag(ScraperApiUrlViperKey.Key(), scraperCmd.Flags().Lookup(ScraperApiUrlViperKey.Flag()))

	scraperCmd.Flags().StringVarP(&scraperOpts.apiKey, ScraperApiKeyViperKey.Flag(), "k", "", "Notifier API key")
	defaultViper.BindPFlag(ScraperApiKeyViperKey.Key(), scraperCmd.Flags().Lookup(ScraperApiKeyViperKey.Flag()))

	scraperCmd.Flags().IntVar(&scraperOpts.interval, "interval", 300, "Interval between each scrape (in seconds)")

	rootCmd.AddCommand(scraperCmd)
}

func (s *ScraperCommand) runScraper() error {
	cfg := buildTokenClientConfiguration()

	year := currentAcademicYear()
	classes := gaps.GetAllClasses(cfg, year)

	ga := gaps.NewGradesAction(cfg, year)
	g, err := ga.FetchGrades()
	if err != nil {
		log.Error("Failed to fetch grades")
		return err
	}

	grades := s.mapGrades(g)
	previousGrades, err := s.readHistory()
	if previousGrades == nil {
		if err != nil {
			log.Error("Failed to read previous grades")
			return err
		}

		log.Info("No previous grades found, overwriting")
		return s.writeHistory(grades)
	}

	diff, _ := diff.Diff(
		previousGrades, grades,
		diff.TagName("diff"),
		diff.DisableStructValues(),
	)

	notifications := make(map[scraperGrade]bool)
	client := notifier.NewClient(s.apiUrl, s.apiKey)

	if len(diff) == 0 {
		log.Info("No changes found")
		return nil
	}

	for _, change := range diff {
		grade := s.resolveGrade(grades, change)
		previous := s.resolveGrade(previousGrades, change)
		if grade == nil || grade.ClassMean == "-" {
			continue
		}

		if notifications[*grade] {
			continue
		}
		notifications[*grade] = true

		nmean, _ := strconv.ParseFloat(grade.ClassMean, 32)
		n := &notifier.ApiGrade{
			Course: grade.Course,
			Class:  s.findClass(grade, classes),
			Name:   grade.Description,
			Mean:   float32(nmean),
		}

		s.logChange(previous, grade, change)

		ctx := context.Background()
		err = client.SendGrade(ctx, n)
		if err != nil {
			log.WithError(err).Error("Failed to send grade")
		}
	}

	return s.writeHistory(grades)
}

func (s *ScraperCommand) mapGrades(grades []*parser.ClassGrades) scraperResult {
	scraperGrades := make(scraperResult)
	for _, class := range grades {
		scraperGrades[class.Name] = make(map[string]*scraperGrade)
		for _, group := range class.GradeGroups {
			for _, grade := range group.Grades {
				scraperGrades[class.Name][grade.Description] = &scraperGrade{
					Course:      class.Name,
					Type:        group.Name,
					Description: grade.Description,
					Date:        grade.Date,
					Weight:      grade.Weight,
					Grade:       grade.Grade,
					ClassMean:   grade.ClassMean,
				}
			}
		}
	}

	return scraperGrades
}

func (s *ScraperCommand) resolveGrade(grades scraperResult, change diff.Change) *scraperGrade {
	if change.Type == diff.DELETE {
		return nil
	}

	return grades[change.Path[0]][change.Path[1]]
}

func (s *ScraperCommand) logChange(previous *scraperGrade, grade *scraperGrade, change diff.Change) {
	if change.Type == diff.CREATE || previous == nil {
		log.WithFields(log.Fields{
			"new-grade":   grade.Grade,
			"new-average": grade.ClassMean,
		}).Infof(
			"NEW [%s] %s: %s (avg: %s).",
			grade.Course,
			grade.Description,
			grade.Grade,
			grade.ClassMean,
		)

		return
	}

	log.WithFields(log.Fields{
		"previous-grade":   previous.Grade,
		"new-grade":        grade.Grade,
		"previous-average": previous.ClassMean,
		"new-average":      grade.ClassMean,
	}).Infof(
		"UPDATED [%s] %s: %s (avg: %s) -> %s (avg: %s).",
		grade.Course,
		grade.Description,
		previous.Grade,
		previous.ClassMean,
		grade.Grade,
		grade.ClassMean,
	)
}

func (s *ScraperCommand) readHistory() (scraperResult, error) {
	var grades scraperResult
	data, err := os.ReadFile(s.historyFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	err = json.Unmarshal(data, &grades)
	return grades, err
}

func (s *ScraperCommand) writeHistory(grades scraperResult) error {
	data, err := json.MarshalIndent(grades, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(s.historyFile, data, 0644)
}

func (s *ScraperCommand) findClass(grade *scraperGrade, classes []string) string {
	re, _ := regexp.Compile(`[\w-]+?-(\w+)-([CL])\d`)

	if grade.Type != "Cours" && grade.Type != "Laboratoire" {
		return grade.Course
	}

	for _, class := range classes {
		if !strings.HasPrefix(class, grade.Course) {
			continue
		}

		matches := re.FindStringSubmatch(class)
		if len(matches) != 3 {
			continue
		}

		className := matches[1]
		classType := matches[2]

		if (grade.Type == "Cours" && classType != "C") || (grade.Type == "Laboratoire" && classType != "L") {
			continue
		}

		return className
	}

	return grade.Course
}
