package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"national-defense-knowledge-quiz/internal/config"
	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
	"national-defense-knowledge-quiz/internal/repository"
)

type ProblemConfig struct {
	Type   string   `json:"type"`
	Text   string   `json:"text"`
	Data   []string `json:"data"`
	Answer string   `json:"answer"`
	Score  int      `json:"score"`
	Active bool     `json:"active"`
}

type PrizeConfig struct {
	Text   string `json:"text"`
	Remain int    `json:"remain"`
}

type ExamConfig struct {
	Exam struct {
		Title       string `json:"title"`
		Intro       string `json:"intro"`
		LimitTime   int    `json:"limit_time"`
		Random      int    `json:"random"`
		LimitNumber int    `json:"limit_number"`
		Active      bool   `json:"active"`
	} `json:"exam"`
	Problems []ProblemConfig `json:"problems"`
	Prizes   []PrizeConfig   `json:"prizes"`
}

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	db.Init(cfg)

	configFile := cfg.ConfigPath
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file %s: %v", configFile, err)
	}

	var examCfg ExamConfig
	if err := json.Unmarshal(data, &examCfg); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	problemRepo := &repository.ProblemRepo{}
	prizeRepo := &repository.PrizeRepo{}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// The config file is the source of truth: wipe the existing rows and
		// recreate them, so re-seeding never leaves duplicate or orphaned data.
		if err := clearExamData(tx); err != nil {
			return err
		}

		exam := model.Exam{
			Title:       examCfg.Exam.Title,
			Intro:       examCfg.Exam.Intro,
			LimitTime:   examCfg.Exam.LimitTime,
			Random:      examCfg.Exam.Random,
			LimitNumber: examCfg.Exam.LimitNumber,
			Active:      examCfg.Exam.Active,
		}
		if err := tx.Create(&exam).Error; err != nil {
			return fmt.Errorf("failed to create exam: %w", err)
		}

		problems := make([]model.Problem, len(examCfg.Problems))
		for i, p := range examCfg.Problems {
			dataBytes, _ := json.Marshal(p.Data)
			problems[i] = model.Problem{
				ExamID: exam.ID,
				Type:   p.Type,
				Text:   p.Text,
				Data:   string(dataBytes),
				Answer: p.Answer,
				Score:  p.Score,
				Active: p.Active,
			}
		}
		if len(problems) > 0 {
			if err := problemRepo.BulkCreate(tx, problems); err != nil {
				return fmt.Errorf("failed to create problems: %w", err)
			}
		}

		prizes := make([]model.Prize, len(examCfg.Prizes))
		for i, p := range examCfg.Prizes {
			prizes[i] = model.Prize{
				ExamID: exam.ID,
				Text:   p.Text,
				Remain: p.Remain,
			}
		}
		if len(prizes) > 0 {
			if err := prizeRepo.BulkCreate(tx, prizes); err != nil {
				return fmt.Errorf("failed to create prizes: %w", err)
			}
		}

		fmt.Printf("seeded exam %d with %d problems and %d prizes\n", exam.ID, len(problems), len(prizes))
		return nil
	})

	if err != nil {
		log.Fatalf("seed failed: %v", err)
	}
}

// clearExamData removes every exam, problem, prize and session row so the seed
// can recreate them from scratch. Sessions reference exams and problems, so
// they are cleared too to avoid dangling references.
func clearExamData(tx *gorm.DB) error {
	if tx.Dialector.Name() == "postgres" {
		// RESTART IDENTITY keeps the exam id at 1, which the frontend expects
		// (it requests /exam/info/?exam=1).
		if err := tx.Exec(
			"TRUNCATE TABLE exam_session_problems, exam_sessions, problems, prizes, exams RESTART IDENTITY",
		).Error; err != nil {
			return fmt.Errorf("failed to clear exam data: %w", err)
		}
		return nil
	}

	for _, table := range []string{"exam_session_problems", "exam_sessions", "problems", "prizes", "exams"} {
		if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}
	// Reset sqlite AUTOINCREMENT counters so the exam id stays 1. Ignore errors
	// on databases that do not maintain a sqlite_sequence table.
	tx.Exec("DELETE FROM sqlite_sequence WHERE name IN ('exams', 'problems', 'prizes', 'exam_sessions')")
	return nil
}
