package game

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// EventType represents different types of events
type EventType string

const (
	EventShinyHunt      EventType = "SHINY_HUNT"      // Increased shiny spawn rates (1/100 instead of 1/4096)
	EventDoubleXP       EventType = "DOUBLE_XP"        // 2x XP from battles
	EventRareEncounter  EventType = "RARE_ENCOUNTER"  // Higher chance of rare/legendary gophers
	EventLuckyDay       EventType = "LUCKY_DAY"       // Better capture rates
	EventStatBoost      EventType = "STAT_BOOST"      // All gophers get 10% stat boost
	EventEvolutionFest  EventType = "EVOLUTION_FEST"  // Faster evolution (reduced level requirements)
)

// Event represents an active event
type Event struct {
	ID          string
	Type        EventType
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Active      bool
}

// EventManager manages active events
type EventManager struct {
	mu      sync.RWMutex
	events  map[string]*Event
	channelID string // Discord channel to announce events
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		events: make(map[string]*Event),
	}
}

// SetAnnouncementChannel sets the Discord channel for event announcements
func (em *EventManager) SetAnnouncementChannel(channelID string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.channelID = channelID
}

// GetAnnouncementChannel returns the announcement channel ID
func (em *EventManager) GetAnnouncementChannel() string {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.channelID
}

// StartEvent starts a new event
func (em *EventManager) StartEvent(eventType EventType, duration time.Duration) *Event {
	em.mu.Lock()
	defer em.mu.Unlock()

	// End any existing event of the same type
	for _, event := range em.events {
		if event.Type == eventType && event.Active {
			event.Active = false
		}
	}

	event := &Event{
		ID:        fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:      eventType,
		Name:      em.getEventName(eventType),
		Description: em.getEventDescription(eventType),
		StartTime: time.Now(),
		EndTime:   time.Now().Add(duration),
		Active:    true,
	}

	em.events[event.ID] = event
	return event
}

// EndEvent ends an event
func (em *EventManager) EndEvent(eventID string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	if event, exists := em.events[eventID]; exists {
		event.Active = false
	}
}

// GetActiveEvents returns all currently active events
func (em *EventManager) GetActiveEvents() []*Event {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var active []*Event
	now := time.Now()
	for _, event := range em.events {
		if event.Active && now.Before(event.EndTime) {
			active = append(active, event)
		} else if event.Active {
			// Event expired, mark as inactive
			event.Active = false
		}
	}

	return active
}

// GetActiveEventByType returns the active event of a specific type
func (em *EventManager) GetActiveEventByType(eventType EventType) *Event {
	em.mu.RLock()
	defer em.mu.RUnlock()

	now := time.Now()
	for _, event := range em.events {
		if event.Type == eventType && event.Active && now.Before(event.EndTime) {
			return event
		}
	}

	return nil
}

// CleanupExpiredEvents removes expired events from memory
func (em *EventManager) CleanupExpiredEvents() {
	em.mu.Lock()
	defer em.mu.Unlock()

	now := time.Now()
	for id, event := range em.events {
		if !event.Active || now.After(event.EndTime) {
			delete(em.events, id)
		}
	}
}

// getEventName returns the display name for an event type
func (em *EventManager) getEventName(eventType EventType) string {
	switch eventType {
	case EventShinyHunt:
		return "✨ Shiny Hunt Event ✨"
	case EventDoubleXP:
		return "⚡ Double XP Event ⚡"
	case EventRareEncounter:
		return "💎 Rare Encounter Event 💎"
	case EventLuckyDay:
		return "🍀 Lucky Day Event 🍀"
	case EventStatBoost:
		return "💪 Stat Boost Event 💪"
	case EventEvolutionFest:
		return "🌟 Evolution Festival 🌟"
	default:
		return "Unknown Event"
	}
}

// getEventDescription returns the description for an event type
func (em *EventManager) getEventDescription(eventType EventType) string {
	switch eventType {
	case EventShinyHunt:
		return "Shiny spawn rates increased to 1/100! Hunt for those rare color variants!"
	case EventDoubleXP:
		return "All battles give 2x XP! Level up your gophers faster!"
	case EventRareEncounter:
		return "Higher chance of encountering RARE, EPIC, and LEGENDARY gophers in the wild!"
	case EventLuckyDay:
		return "Capture rates are significantly improved! Catch those gophers!"
	case EventStatBoost:
		return "All gophers get a 10% stat boost! Your team is stronger!"
	case EventEvolutionFest:
		return "Evolution requirements reduced! Your gophers evolve faster!"
	default:
		return "An event is active!"
	}
}

// GetShinyRate returns the current shiny spawn rate (affected by events)
func (em *EventManager) GetShinyRate() float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Check for shiny hunt event
	if em.GetActiveEventByType(EventShinyHunt) != nil {
		return 1.0 / 100.0 // 1/100 during event
	}

	return 1.0 / 4096.0 // Normal rate
}

// GetXPMultiplier returns the XP multiplier (affected by events)
func (em *EventManager) GetXPMultiplier() float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.GetActiveEventByType(EventDoubleXP) != nil {
		return 2.0
	}

	return 1.0
}

// GetRarityBoost returns rarity encounter boost (affected by events)
func (em *EventManager) GetRarityBoost() float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.GetActiveEventByType(EventRareEncounter) != nil {
		return 2.0 // Double chance of rare encounters
	}

	return 1.0
}

// GetCaptureRateMultiplier returns capture rate multiplier (affected by events)
func (em *EventManager) GetCaptureRateMultiplier() float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.GetActiveEventByType(EventLuckyDay) != nil {
		return 1.5 // 50% better capture rates
	}

	return 1.0
}

// GetStatBoostMultiplier returns stat boost multiplier (affected by events)
func (em *EventManager) GetStatBoostMultiplier() float64 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.GetActiveEventByType(EventStatBoost) != nil {
		return 1.10 // 10% stat boost
	}

	return 1.0
}

// GetEvolutionLevelReduction returns how many levels are reduced for evolution (affected by events)
func (em *EventManager) GetEvolutionLevelReduction() int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.GetActiveEventByType(EventEvolutionFest) != nil {
		return 5 // Reduce evolution requirement by 5 levels
	}

	return 0
}

// GetRandomEventType returns a random event type for automatic events
func GetRandomEventType() EventType {
	events := []EventType{
		EventShinyHunt,
		EventDoubleXP,
		EventRareEncounter,
		EventLuckyDay,
		EventStatBoost,
		EventEvolutionFest,
	}
	return events[rand.Intn(len(events))]
}

// StartRandomEvent starts a random event with the given duration
func (em *EventManager) StartRandomEvent(duration time.Duration) *Event {
	eventType := GetRandomEventType()
	return em.StartEvent(eventType, duration)
}

