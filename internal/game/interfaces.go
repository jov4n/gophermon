package game

import (
	"gophermon-bot/internal/storage"
)

// Shared interfaces for game services

// TrainerRepoInterface defines methods needed from trainer repository
type TrainerRepoInterface interface {
	GetCurrency(trainerID string) (int, error)
	AddCurrency(trainerID string, amount int) error
	RemoveCurrency(trainerID string, amount int) error
}

// ItemRepoInterface defines methods needed from item repository
type ItemRepoInterface interface {
	AddItem(trainerID, itemType string, quantity int) error
	UseItem(trainerID, itemType string, quantity int) error
	GetItemQuantity(trainerID, itemType string) (int, error)
	GetItems(trainerID string) ([]*storage.Item, error)
}

// AchievementRepoInterface defines methods needed from achievement repository
type AchievementRepoInterface interface {
	GetOrCreate(trainerID, achievementType string) (*storage.Achievement, error)
	UpdateProgress(trainerID, achievementType string, progress int) error
	Complete(trainerID, achievementType string) error
	GetAchievements(trainerID string) ([]*storage.Achievement, error)
}

// QuestRepoInterface defines methods needed from quest repository
type QuestRepoInterface interface {
	Create(quest *storage.Quest) error
	GetActiveQuests(trainerID string) ([]*storage.Quest, error)
	GetQuestByType(trainerID, questType, questName string) (*storage.Quest, error)
	UpdateProgress(questID string, progress int) error
	Complete(questID string) error
	CleanupExpired() error
}

