package services

import (
	"context"
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type MessageService interface {
	SendMessage(ctx context.Context, senderID, conversationID, content string) (*models.Message, error)
	CreateConversation(ctx context.Context, senderID, recipientID, subject, jobID string) (*models.Conversation, error)
	GetConversation(ctx context.Context, conversationID, userID string) (*models.Conversation, error)
	ListConversations(ctx context.Context, userID string, page, limit int) (*models.ConversationListResponse, error)
	ListMessages(ctx context.Context, conversationID, userID string, page, limit int) (*models.MessageListResponse, error)
	MarkConversationRead(ctx context.Context, conversationID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type MessageServiceImpl struct {
	messageRepo     repository.MessageRepository
	notificationSvc NotificationService
	websocketHub    *WebSocketHub
}

func NewMessageService(messageRepo repository.MessageRepository, notificationSvc NotificationService) MessageService {
	return &MessageServiceImpl{
		messageRepo:     messageRepo,
		notificationSvc: notificationSvc,
		websocketHub:    GetHub(),
	}
}

func (s *MessageServiceImpl) CreateConversation(ctx context.Context, senderID, recipientID, subject, jobID string) (*models.Conversation, error) {
	if subject == "" {
		subject = "Conversation"
	}
	conv := &models.Conversation{
		Subject: subject,
	}
	if jobID != "" {
		conv.JobID = &jobID
	}
	if err := s.messageRepo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	for _, uid := range []string{senderID, recipientID} {
		participant := &models.ConversationParticipant{
			ConversationID: conv.ID,
			UserID:         uid,
			LastReadAt:     time.Now(),
		}
		if err := s.messageRepo.AddParticipant(ctx, participant); err != nil {
			return nil, err
		}
	}

	return conv, nil
}

func (s *MessageServiceImpl) SendMessage(ctx context.Context, senderID, conversationID, content string) (*models.Message, error) {
	if content == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	conv, err := s.messageRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	isParticipant := false
	for _, p := range conv.Participants {
		if p.UserID == senderID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, fmt.Errorf("not a participant of this conversation")
	}

	msg := &models.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
	}
	if err := s.messageRepo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	msg.Sender = &models.User{ID: senderID}

	go s.messageRepo.MarkConversationRead(context.Background(), conversationID, senderID)

	go func() {
		participants, err := s.messageRepo.GetConversationParticipants(context.Background(), conversationID)
		if err != nil {
			return
		}
		for _, p := range participants {
			if p.UserID == senderID {
				continue
			}
			s.notificationSvc.NotifyMessageReceived(context.Background(), p.UserID, senderID, content)
			s.websocketHub.SendToUser(p.UserID, models.WebSocketMessage{
				Type:      "message",
				Data:      msg,
				Timestamp: time.Now(),
			})
		}
	}()

	return msg, nil
}

func (s *MessageServiceImpl) GetConversation(ctx context.Context, conversationID, userID string) (*models.Conversation, error) {
	conv, err := s.messageRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isParticipant := false
	for _, p := range conv.Participants {
		if p.UserID == userID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, fmt.Errorf("not a participant of this conversation")
	}

	lastMsg, _ := s.messageRepo.GetLastMessage(ctx, conversationID)
	conv.LastMessage = lastMsg
	return conv, nil
}

func (s *MessageServiceImpl) ListConversations(ctx context.Context, userID string, page, limit int) (*models.ConversationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	convs, total, err := s.messageRepo.ListUserConversations(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	unreadCount, _ := s.messageRepo.GetUnreadCount(ctx, userID)

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &models.ConversationListResponse{
		Conversations: convs,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	}, nil
}

func (s *MessageServiceImpl) ListMessages(ctx context.Context, conversationID, userID string, page, limit int) (*models.MessageListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	_, err := s.messageRepo.GetParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("not a participant of this conversation")
	}

	msgs, total, err := s.messageRepo.ListMessages(ctx, conversationID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &models.MessageListResponse{
		Messages:   msgs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *MessageServiceImpl) MarkConversationRead(ctx context.Context, conversationID, userID string) error {
	return s.messageRepo.MarkConversationRead(ctx, conversationID, userID)
}

func (s *MessageServiceImpl) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.messageRepo.GetUnreadCount(ctx, userID)
}
