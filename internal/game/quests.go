package game

import (
	"time"
	"gophermon-bot/internal/storage"
)

// Quest types
const (
	QuestTypeDaily  = "DAILY"
	QuestTypeWeekly = "WEEKLY"
)

// Quest names
const (
	QuestWinBattles     = "Win 3 Battles"
	QuestCatchGophers   = "Catch 5 Gophers"
	QuestEvolveGopher   = "Evolve 1 Gopher"
	QuestCatchShiny     = "Catch 1 Shiny"
	QuestWin10Battles   = "Win 10 Battles"
	QuestCatch20Gophers = "Catch 20 Gophers"
)

type QuestService struct {
	questRepo   QuestRepoInterface
	trainerRepo TrainerRepoInterface
}

func NewQuestService(questRepo QuestRepoInterface, trainerRepo TrainerRepoInterface) *QuestService {
	return &QuestService{
		questRepo:   questRepo,
		trainerRepo: trainerRepo,
	}
}

func (s *QuestService) GenerateDailyQuests(trainerID string) error {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	// Set to midnight tomorrow
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	quests := []struct {
		Name        string
		Description string
		Target      int
		Reward      int
	}{
		{QuestWinBattles, "Win 3 battles today", 3, 100},
		{QuestCatchGophers, "Catch 5 gophers today", 5, 150},
		{QuestEvolveGopher, "Evolve 1 gopher today", 1, 200},
	}

		for _, q := range quests {
		// Check if quest already exists
		existing, _ := s.questRepo.GetQuestByType(trainerID, QuestTypeDaily, q.Name)
		if existing != nil {
			continue // Quest already exists
		}

		// Create quest
		quest := &storage.Quest{
			TrainerID:      trainerID,
			QuestType:      QuestTypeDaily,
			QuestName:      q.Name,
			Description:    q.Description,
			TargetValue:    q.Target,
			CurrentProgress: 0,
			RewardCurrency: q.Reward,
			RewardXP:      0,
			Completed:     false,
			ExpiresAt:     tomorrow,
		}
		if err := s.questRepo.Create(quest); err != nil {
			// Log error but continue with other quests
			// Return error would prevent other quests from being created
			// In production, you might want to collect errors and return them
			continue
		}
	}

	return nil
}

func (s *QuestService) UpdateQuestProgress(trainerID, questName string, progress int) error {
	// Try daily first
	quest, err := s.questRepo.GetQuestByType(trainerID, QuestTypeDaily, questName)
	if err == nil && quest != nil {
		s.questRepo.UpdateProgress(quest.ID, progress)
		// Check completion
		if quest.CurrentProgress+progress >= quest.TargetValue && !quest.Completed {
			s.questRepo.Complete(quest.ID)
			// Give rewards
			if quest.RewardCurrency > 0 {
				s.trainerRepo.AddCurrency(trainerID, quest.RewardCurrency)
			}
		}
		return nil
	}

	// Try weekly
	quest, err = s.questRepo.GetQuestByType(trainerID, QuestTypeWeekly, questName)
	if err == nil && quest != nil {
		s.questRepo.UpdateProgress(quest.ID, progress)
		// Check completion
		if quest.CurrentProgress+progress >= quest.TargetValue && !quest.Completed {
			s.questRepo.Complete(quest.ID)
			// Give rewards
			if quest.RewardCurrency > 0 {
				s.trainerRepo.AddCurrency(trainerID, quest.RewardCurrency)
			}
		}
		return nil
	}

	return nil
}

