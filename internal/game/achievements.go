package game

// Achievement types
const (
	AchievementFirstCatch      = "FIRST_CATCH"
	AchievementShinyHunter     = "SHINY_HUNTER"      // Catch 10 shiny gophers
	AchievementEvolutionMaster = "EVOLUTION_MASTER"  // Evolve 50 gophers
	AchievementBattleVeteran   = "BATTLE_VETERAN"    // Win 100 battles
	AchievementLegendaryCollector = "LEGENDARY_COLLECTOR" // Own 5 legendary gophers
	AchievementCatchMaster     = "CATCH_MASTER"       // Catch 100 gophers
	AchievementLevelMaster     = "LEVEL_MASTER"       // Reach level 50 with a gopher
)

// Achievement requirements
var AchievementRequirements = map[string]int{
	AchievementFirstCatch:      1,
	AchievementShinyHunter:     10,
	AchievementEvolutionMaster:  50,
	AchievementBattleVeteran:   100,
	AchievementLegendaryCollector: 5,
	AchievementCatchMaster:     100,
	AchievementLevelMaster:     50,
}

// Achievement rewards (currency)
var AchievementRewards = map[string]int{
	AchievementFirstCatch:      50,
	AchievementShinyHunter:     500,
	AchievementEvolutionMaster:  1000,
	AchievementBattleVeteran:   1000,
	AchievementLegendaryCollector: 2000,
	AchievementCatchMaster:     500,
	AchievementLevelMaster:     1000,
}

type AchievementService struct {
	achievementRepo AchievementRepoInterface
	trainerRepo     TrainerRepoInterface
}

func NewAchievementService(achievementRepo AchievementRepoInterface, trainerRepo TrainerRepoInterface) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		trainerRepo:     trainerRepo,
	}
}

func (s *AchievementService) CheckAndUpdateAchievement(trainerID, achievementType string, progress int) (completed bool, reward int, err error) {
	ach, err := s.achievementRepo.GetOrCreate(trainerID, achievementType)
	if err != nil {
		return false, 0, err
	}

	if ach.Completed {
		return false, 0, nil // Already completed
	}

	// Update progress
	if err := s.achievementRepo.UpdateProgress(trainerID, achievementType, progress); err != nil {
		return false, 0, err
	}

	// Get updated achievement to check if completed
	ach, err = s.achievementRepo.GetOrCreate(trainerID, achievementType)
	if err != nil {
		return false, 0, err
	}

	// Check completion
	requirement := AchievementRequirements[achievementType]
	if ach.Progress >= requirement && !ach.Completed {
		// Complete the achievement
		if err := s.CompleteAchievement(trainerID, achievementType); err != nil {
			return false, 0, err
		}
		return true, AchievementRewards[achievementType], nil
	}

	return false, 0, nil
}

func (s *AchievementService) CompleteAchievement(trainerID, achievementType string) error {
	if err := s.achievementRepo.Complete(trainerID, achievementType); err != nil {
		return err
	}

	// Give reward
	reward := AchievementRewards[achievementType]
	if reward > 0 {
		if err := s.trainerRepo.AddCurrency(trainerID, reward); err != nil {
			// Log error but don't fail
			return nil
		}
	}

	return nil
}

