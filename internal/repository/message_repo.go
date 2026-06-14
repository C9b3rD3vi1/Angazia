package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type MessageRepository interface {
	CreateConversation(ctx context.Context, conv *models.Conversation) error
	GetConversation(ctx context.Context, id string) (*models.Conversation, error)
	ListUserConversations(ctx context.Context, userID string, page, limit int) ([]*models.Conversation, int64, error)
	AddParticipant(ctx context.Context, participant *models.ConversationParticipant) error
	GetParticipant(ctx context.Context, conversationID, userID string) (*models.ConversationParticipant, error)
	GetConversationParticipants(ctx context.Context, conversationID string) ([]*models.ConversationParticipant, error)

	CreateMessage(ctx context.Context, msg *models.Message) error
	GetMessage(ctx context.Context, id string) (*models.Message, error)
	ListMessages(ctx context.Context, conversationID string, page, limit int) ([]*models.Message, int64, error)
	MarkConversationRead(ctx context.Context, conversationID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	GetLastMessage(ctx context.Context, conversationID string) (*models.Message, error)
}

type MessageRepositoryImpl struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &MessageRepositoryImpl{db: db}
}

func (r *MessageRepositoryImpl) CreateConversation(ctx context.Context, conv *models.Conversation) error {
	conv.ID = uuid.New().String()
	conv.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *MessageRepositoryImpl) GetConversation(ctx context.Context, id string) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.WithContext(ctx).Preload("Participants").First(&conv, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *MessageRepositoryImpl) ListUserConversations(ctx context.Context, userID string, page, limit int) ([]*models.Conversation, int64, error) {
	var convs []*models.Conversation
	var total int64

	sub := r.db.WithContext(ctx).Model(&models.ConversationParticipant{}).
		Select("conversation_id").
		Where("user_id = ? AND is_archived = ?", userID, false)

	query := r.db.WithContext(ctx).Model(&models.Conversation{}).
		Where("id IN (?)", sub)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Participants", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).
		Order("updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&convs).Error
	if err != nil {
		return nil, 0, err
	}

	for _, c := range convs {
		lastMsg, _ := r.GetLastMessage(ctx, c.ID)
		if lastMsg != nil {
			c.LastMessage = lastMsg
		}
	}

	return convs, total, nil
}

func (r *MessageRepositoryImpl) AddParticipant(ctx context.Context, participant *models.ConversationParticipant) error {
	participant.ID = uuid.New().String()
	participant.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(participant).Error
}

func (r *MessageRepositoryImpl) GetParticipant(ctx context.Context, conversationID, userID string) (*models.ConversationParticipant, error) {
	var p models.ConversationParticipant
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *MessageRepositoryImpl) GetConversationParticipants(ctx context.Context, conversationID string) ([]*models.ConversationParticipant, error) {
	var participants []*models.ConversationParticipant
	err := r.db.WithContext(ctx).Preload("User").
		Where("conversation_id = ?", conversationID).
		Find(&participants).Error
	return participants, err
}

func (r *MessageRepositoryImpl) CreateMessage(ctx context.Context, msg *models.Message) error {
	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *MessageRepositoryImpl) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	var msg models.Message
	err := r.db.WithContext(ctx).Preload("Sender").First(&msg, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepositoryImpl) ListMessages(ctx context.Context, conversationID string, page, limit int) ([]*models.Message, int64, error) {
	var msgs []*models.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Message{}).
		Where("conversation_id = ?", conversationID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Sender").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&msgs).Error
	return msgs, total, err
}

func (r *MessageRepositoryImpl) MarkConversationRead(ctx context.Context, conversationID, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_at", now).Error
}

func (r *MessageRepositoryImpl) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM messages
		JOIN conversation_participants ON conversation_participants.conversation_id = messages.conversation_id
			AND conversation_participants.user_id = ?
		WHERE messages.sender_id != ?
			AND messages.created_at > conversation_participants.last_read_at
	`, userID, userID).Scan(&count).Error
	return int(count), err
}

func (r *MessageRepositoryImpl) GetLastMessage(ctx context.Context, conversationID string) (*models.Message, error) {
	var msg models.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
